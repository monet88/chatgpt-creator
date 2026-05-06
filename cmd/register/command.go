package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/monet88/chatgpt-creator/internal/config"
	"github.com/monet88/chatgpt-creator/internal/email"
	proxypool "github.com/monet88/chatgpt-creator/internal/proxy"
	"github.com/monet88/chatgpt-creator/internal/register"
	"github.com/spf13/cobra"
)

const (
	exitCodeValidation = 2
	exitCodeConfig     = 3
	exitCodeRuntime    = 4
)

var runBatchForCLI = register.RunBatchForCLI
var runBatchWithProviders = register.RunBatchForCLIWithProviders
var withDiagnosticWriter = register.WithDiagnosticWriter

type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string {
	return e.err.Error()
}

func executeWithIO(in io.Reader, out, errOut io.Writer) int {
	cmd := newRegisterCommand(in, out, errOut)
	if err := cmd.Execute(); err != nil {
		var ee *exitError
		if errors.As(err, &ee) {
			fmt.Fprintln(errOut, ee.Error())
			return ee.code
		}
		fmt.Fprintln(errOut, err.Error())
		return exitCodeRuntime
	}
	return 0
}

func newRegisterCommand(in io.Reader, out, errOut io.Writer) *cobra.Command {
	var (
		configPath    string
		total         int
		workers       int
		proxy         string
		proxyList     string
		proxyCooldown int
		outputFile    string
		password      string
		domain        string
		pacing        string
		jsonMode      bool
		interactive   bool
		// ViOTP flags
		viOTPToken     string
		viOTPServiceID int
		// IMAP flags
		imapHost     string
		imapPort     int
		imapUser     string
		imapPassword string
		imapTLS      bool
		// Codex flags
		codexEnabled bool
		codexOutput  string
	)

	cmd := &cobra.Command{
		Use:           "register",
		Short:         "Register ChatGPT accounts",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				return &exitError{code: exitCodeConfig, err: fmt.Errorf("error loading config: %w", err)}
			}

			effectiveProxy := cfg.Proxy
			if cmd.Flags().Changed("proxy") {
				effectiveProxy = proxy
			}

			effectiveProxyList := cfg.ProxyList
			if cmd.Flags().Changed("proxy-list") {
				effectiveProxyList = proxyList
			}

			if effectiveProxy != "" && effectiveProxyList != "" {
				return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: --proxy and --proxy-list are mutually exclusive")}
			}

			effectiveOutput := cfg.OutputFile
			if cmd.Flags().Changed("output") {
				effectiveOutput = outputFile
			}

			effectivePassword := cfg.DefaultPassword
			if cmd.Flags().Changed("password") {
				effectivePassword = password
			}

			effectiveDomain := cfg.DefaultDomain
			if cmd.Flags().Changed("domain") {
				effectiveDomain = domain
			}

			hasActionableFlags := cmd.Flags().Changed("total") ||
				cmd.Flags().Changed("workers") ||
				cmd.Flags().Changed("proxy") ||
				cmd.Flags().Changed("proxy-list") ||
				cmd.Flags().Changed("proxy-cooldown") ||
				cmd.Flags().Changed("output") ||
				cmd.Flags().Changed("password") ||
				cmd.Flags().Changed("domain") ||
				cmd.Flags().Changed("pacing") ||
				cmd.Flags().Changed("json") ||
				cmd.Flags().Changed("imap-host") ||
				cmd.Flags().Changed("imap-port") ||
				cmd.Flags().Changed("imap-user") ||
				cmd.Flags().Changed("imap-password") ||
				cmd.Flags().Changed("imap-tls") ||
				cmd.Flags().Changed("viotp-token") ||
				cmd.Flags().Changed("viotp-service-id") ||
				cmd.Flags().Changed("codex") ||
				cmd.Flags().Changed("codex-output")
			if interactive || !hasActionableFlags {
				return runInteractive(in, out, errOut, cfg, effectiveOutput)
			}

			effectiveViOTPToken := cfg.ViOTPToken
			if cmd.Flags().Changed("viotp-token") {
				effectiveViOTPToken = viOTPToken
			}
			effectiveViOTPServiceID := cfg.ViOTPServiceID
			if cmd.Flags().Changed("viotp-service-id") {
				effectiveViOTPServiceID = viOTPServiceID
			}
			if effectiveViOTPToken != "" || effectiveViOTPServiceID > 0 {
				return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: viotp phone challenge automation is not supported in safe mode")}
			}

			effectiveCodexEnabled := cfg.CodexEnabled
			if cmd.Flags().Changed("codex") {
				effectiveCodexEnabled = codexEnabled
			}
			if effectiveCodexEnabled {
				return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: codex extraction is not supported in safe mode")}
			}
			if cmd.Flags().Changed("codex-output") || cfg.CodexOutput != config.DefaultCodexOutput {
				return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: codex output is not supported in safe mode")}
			}

			var proxyPool *proxypool.RoundRobinPool
			if effectiveProxyList != "" {
				proxies, loadErr := proxypool.LoadProxies(effectiveProxyList)
				if loadErr != nil {
					return &exitError{code: exitCodeConfig, err: fmt.Errorf("config error: %w", loadErr)}
				}
				cooldownSec := proxyCooldown
				if cfg.ProxyCooldown > 0 && !cmd.Flags().Changed("proxy-cooldown") {
					cooldownSec = cfg.ProxyCooldown
				}
				pool, poolErr := proxypool.NewRoundRobinPool(proxies, time.Duration(cooldownSec)*time.Second)
				if poolErr != nil {
					return &exitError{code: exitCodeConfig, err: fmt.Errorf("config error: %w", poolErr)}
				}
				proxyPool = pool
			} else if effectiveProxy != "" {
				proxyPool = proxypool.NewSinglePool(effectiveProxy)
			}

			effectiveIMAPHost := cfg.IMAPHost
			if cmd.Flags().Changed("imap-host") {
				effectiveIMAPHost = imapHost
			}
			effectiveIMAPUser := cfg.IMAPUser
			if cmd.Flags().Changed("imap-user") {
				effectiveIMAPUser = imapUser
			}
			effectiveIMAPPassword := cfg.IMAPPassword
			if cmd.Flags().Changed("imap-password") {
				effectiveIMAPPassword = imapPassword
			}
			effectiveIMAPPort := cfg.IMAPPort
			if cmd.Flags().Changed("imap-port") {
				effectiveIMAPPort = imapPort
			}
			effectiveIMAPTLS := cfg.IMAPUseTLS
			if cmd.Flags().Changed("imap-tls") {
				effectiveIMAPTLS = imapTLS
			}

			var otpProvider email.OTPProvider
			if effectiveIMAPHost != "" && effectiveIMAPUser != "" && effectiveIMAPPassword != "" {
				pooler, imapErr := email.NewIMAPPooler(email.IMAPConfig{
					Host:     effectiveIMAPHost,
					Port:     effectiveIMAPPort,
					User:     effectiveIMAPUser,
					Password: effectiveIMAPPassword,
					UseTLS:   effectiveIMAPTLS,
				})
				if imapErr != nil {
					return &exitError{code: exitCodeConfig, err: fmt.Errorf("config error: IMAP connection failed: %w", imapErr)}
				}
				defer pooler.Close()
				otpProvider = pooler
			}

			if total <= 0 {
				return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: --total must be greater than 0")}
			}
			if workers <= 0 {
				return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: --workers must be greater than 0")}
			}
			if effectivePassword != "" && len(effectivePassword) < 12 {
				return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: password must be at least 12 characters")}
			}
			if strings.TrimSpace(effectiveOutput) == "" {
				return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: --output must not be empty")}
			}

			effectivePacing := cfg.Pacing
			if cmd.Flags().Changed("pacing") {
				effectivePacing = pacing
			}
			if effectivePacing == "" {
				effectivePacing = "human"
			}
			pacingProfile, pacingErr := register.ParsePacingProfile(effectivePacing)
			if pacingErr != nil {
				return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: %w", pacingErr)}
			}

			providers := register.ProviderOptions{
				OTPProvider: otpProvider,
			}
			if proxyPool != nil {
				providers.ProxyPool = proxyPool
			}

			if jsonMode {
				var result register.BatchResult
				withDiagnosticWriter(errOut, func() {
					opts := register.DefaultBatchOptionsForCLI(total)
					opts.PacingProfile = pacingProfile
					result = runBatchWithProviders(context.Background(), total, effectiveOutput, workers, effectiveProxy, effectivePassword, effectiveDomain, opts, providers)
				})
				if err := json.NewEncoder(out).Encode(result); err != nil {
					return &exitError{code: exitCodeRuntime, err: fmt.Errorf("runtime error: failed to encode json result: %w", err)}
				}
				if result.Success < int64(result.Target) {
					return &exitError{code: exitCodeRuntime, err: fmt.Errorf("runtime error: target not reached (%d/%d), stop_reason=%s", result.Success, result.Target, result.StopReason)}
				}
				return nil
			}

			opts := register.DefaultBatchOptionsForCLI(total)
			opts.PacingProfile = pacingProfile
			result := runBatchWithProviders(context.Background(), total, effectiveOutput, workers, effectiveProxy, effectivePassword, effectiveDomain, opts, providers)
			if result.Success < int64(result.Target) {
				return &exitError{code: exitCodeRuntime, err: fmt.Errorf("runtime error: target not reached (%d/%d), stop_reason=%s", result.Success, result.Target, result.StopReason)}
			}
			return nil
		},
	}

	cmd.SetIn(in)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	cmd.Flags().StringVar(&configPath, "config", config.DefaultConfigPath(), "Path to config file")
	cmd.Flags().IntVar(&total, "total", 0, "Total accounts to register")
	cmd.Flags().IntVar(&workers, "workers", 3, "Max concurrent workers")
	cmd.Flags().StringVar(&proxy, "proxy", "", "Proxy URL")
	cmd.Flags().StringVar(&proxyList, "proxy-list", "", "Path to proxy list file (one URL per line)")
	cmd.Flags().IntVar(&proxyCooldown, "proxy-cooldown", 300, "Proxy cooldown in seconds after failure")
	cmd.Flags().StringVar(&outputFile, "output", "", "Output file path")
	cmd.Flags().StringVar(&password, "password", "", "Default password")
	cmd.Flags().StringVar(&domain, "domain", "", "Default email domain")
	cmd.Flags().StringVar(&pacing, "pacing", "human", "Pacing profile: none, fast, human, slow")
	cmd.Flags().BoolVar(&jsonMode, "json", false, "Emit machine-readable summary")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Force interactive prompt mode")
	// ViOTP flags
	cmd.Flags().StringVar(&viOTPToken, "viotp-token", "", "ViOTP API token for SMS verification")
	cmd.Flags().IntVar(&viOTPServiceID, "viotp-service-id", 0, "ViOTP service ID for OpenAI")
	// IMAP flags
	cmd.Flags().StringVar(&imapHost, "imap-host", "", "IMAP server hostname for catch-all email")
	cmd.Flags().IntVar(&imapPort, "imap-port", 993, "IMAP server port")
	cmd.Flags().StringVar(&imapUser, "imap-user", "", "IMAP username")
	cmd.Flags().StringVar(&imapPassword, "imap-password", "", "IMAP password")
	cmd.Flags().BoolVar(&imapTLS, "imap-tls", true, "Use TLS for IMAP connection")
	// Codex flags
	cmd.Flags().BoolVar(&codexEnabled, "codex", false, "Enable post-registration Codex token extraction")
	cmd.Flags().StringVar(&codexOutput, "codex-output", config.DefaultCodexOutput, "Unsupported in safe mode; Codex token output path")

	return cmd
}

func runInteractive(in io.Reader, out, errOut io.Writer, cfg *config.Config, outputFile string) error {
	printBanner(out)

	reader := bufio.NewReader(in)

	proxy := cfg.Proxy
	if cfg.Proxy == "" {
		fmt.Fprint(out, "Proxy (enter to skip): ")
		proxyInput, _ := reader.ReadString('\n')
		proxy = strings.TrimSpace(proxyInput)
	}

	fmt.Fprint(out, "Total accounts to register: ")
	totalInput, _ := reader.ReadString('\n')
	totalInput = strings.TrimSpace(totalInput)
	if totalInput == "" {
		return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: total accounts is required")}
	}

	totalAccounts, err := strconv.Atoi(totalInput)
	if err != nil || totalAccounts <= 0 {
		return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: invalid total accounts %q", totalInput)}
	}

	defaultWorkers := 3
	fmt.Fprintf(out, "Max concurrent workers (default: %d): ", defaultWorkers)
	workersInput, _ := reader.ReadString('\n')
	workersInput = strings.TrimSpace(workersInput)

	maxWorkers := defaultWorkers
	if workersInput != "" {
		parsedWorkers, parseErr := strconv.Atoi(workersInput)
		if parseErr != nil || parsedWorkers <= 0 {
			return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: invalid worker count %q", workersInput)}
		}
		maxWorkers = parsedWorkers
	}

	defaultPassword := cfg.DefaultPassword
	if cfg.DefaultPassword == "" {
		fmt.Fprint(out, "Default password (current: (random), press Enter to use, or enter new): ")
		pwInput, _ := reader.ReadString('\n')
		pwInput = strings.TrimSpace(pwInput)
		if pwInput != "" {
			if len(pwInput) < 12 {
				return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: password must be at least 12 characters")}
			}
			defaultPassword = pwInput
		}
	}

	defaultDomain := cfg.DefaultDomain
	if cfg.DefaultDomain == "" {
		fmt.Fprint(out, "Default domain (current: (random from generator.email), press Enter to use, or enter new): ")
		domainInput, _ := reader.ReadString('\n')
		domainInput = strings.TrimSpace(domainInput)
		if domainInput != "" {
			defaultDomain = domainInput
		}
	}

	pacingStr := cfg.Pacing
	if pacingStr == "" {
		pacingStr = "human"
	}
	fmt.Fprintf(out, "Pacing profile (current: %s, options: none/fast/human/slow): ", pacingStr)
	pacingInput, _ := reader.ReadString('\n')
	pacingInput = strings.TrimSpace(pacingInput)
	if pacingInput != "" {
		pacingStr = pacingInput
	}
	pacingProfile, pacingErr := register.ParsePacingProfile(pacingStr)
	if pacingErr != nil {
		return &exitError{code: exitCodeValidation, err: fmt.Errorf("validation error: %w", pacingErr)}
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "-------------------------------------------")
	fmt.Fprintln(out, "Configuration:")
	fmt.Fprintf(out, "  Proxy:          %s\n", register.RedactProxy(proxy))
	fmt.Fprintf(out, "  Total Accounts: %d\n", totalAccounts)
	fmt.Fprintf(out, "  Max Workers:    %d\n", maxWorkers)
	fmt.Fprintf(out, "  Password:       %s\n", register.RedactPassword(defaultPassword))
	if defaultDomain != "" {
		fmt.Fprintf(out, "  Domain:         %s\n", defaultDomain)
	} else {
		fmt.Fprintln(out, "  Domain:         (random)")
	}
	fmt.Fprintf(out, "  Pacing:         %s\n", pacingProfile)
	fmt.Fprintf(out, "  Output File:    %s\n", outputFile)
	fmt.Fprintln(out, "-------------------------------------------")
	fmt.Fprintln(out)

	opts := register.DefaultBatchOptionsForCLI(totalAccounts)
	opts.PacingProfile = pacingProfile
	_ = runBatchForCLI(context.Background(), totalAccounts, outputFile, maxWorkers, proxy, defaultPassword, defaultDomain, opts)
	return nil
}

func printBanner(out io.Writer) {
	banner := `
   _____ _           _    _____ _____ _______
  / ____| |         | |  / ____|  __ \__   __|
 | |    | |__   __ _| |_| |  __| |__) | | |
 | |    | '_ \ / _` + "`" + ` | __| | |_ |  ___/  | |
 | |____| | | | (_| | |_| |__| | |      | |
  \_____|_| |_|\__,_|\__|\_____|_|      |_|

      ChatGPT Account Registration Bot
               by @verssache
`
	fmt.Fprintln(out, banner)
}

func runMain() int {
	return executeWithIO(os.Stdin, os.Stdout, os.Stderr)
}
