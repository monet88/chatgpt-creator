package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/verssache/chatgpt-creator/internal/config"
	"github.com/verssache/chatgpt-creator/internal/register"
)

const (
	exitCodeValidation = 2
	exitCodeConfig     = 3
	exitCodeRuntime    = 4
)

var runBatchForCLI = register.RunBatchForCLI
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
		configPath   string
		total        int
		workers      int
		proxy        string
		outputFile   string
		password     string
		domain       string
		jsonMode     bool
		interactive  bool
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

			hasActionableFlags := cmd.Flags().Changed("total") || cmd.Flags().Changed("workers") || cmd.Flags().Changed("proxy") || cmd.Flags().Changed("output") || cmd.Flags().Changed("password") || cmd.Flags().Changed("domain") || cmd.Flags().Changed("json")
			if interactive || !hasActionableFlags {
				return runInteractive(in, out, errOut, cfg, effectiveOutput)
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

			if jsonMode {
				var result register.BatchResult
				withDiagnosticWriter(errOut, func() {
					result = runBatchForCLI(context.Background(), total, effectiveOutput, workers, effectiveProxy, effectivePassword, effectiveDomain)
				})
				if err := json.NewEncoder(out).Encode(result); err != nil {
					return &exitError{code: exitCodeRuntime, err: fmt.Errorf("runtime error: failed to encode json result: %w", err)}
				}
				if result.Success < int64(result.Target) {
					return &exitError{code: exitCodeRuntime, err: fmt.Errorf("runtime error: target not reached (%d/%d), stop_reason=%s", result.Success, result.Target, result.StopReason)}
				}
				return nil
			}

			result := runBatchForCLI(context.Background(), total, effectiveOutput, workers, effectiveProxy, effectivePassword, effectiveDomain)
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
	cmd.Flags().StringVar(&outputFile, "output", "", "Output file path")
	cmd.Flags().StringVar(&password, "password", "", "Default password")
	cmd.Flags().StringVar(&domain, "domain", "", "Default email domain")
	cmd.Flags().BoolVar(&jsonMode, "json", false, "Emit machine-readable summary")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Force interactive prompt mode")

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

	fmt.Fprintln(out)
	fmt.Fprintln(out, "-------------------------------------------")
	fmt.Fprintln(out, "Configuration:")
	fmt.Fprintf(out, "  Proxy:          %s\n", redactProxyForDisplay(proxy))
	fmt.Fprintf(out, "  Total Accounts: %d\n", totalAccounts)
	fmt.Fprintf(out, "  Max Workers:    %d\n", maxWorkers)
	fmt.Fprintf(out, "  Password:       %s\n", redactPasswordForDisplay(defaultPassword))
	if defaultDomain != "" {
		fmt.Fprintf(out, "  Domain:         %s\n", defaultDomain)
	} else {
		fmt.Fprintln(out, "  Domain:         (random)")
	}
	fmt.Fprintf(out, "  Output File:    %s\n", outputFile)
	fmt.Fprintln(out, "-------------------------------------------")
	fmt.Fprintln(out)

	_ = runBatchForCLI(context.Background(), totalAccounts, outputFile, maxWorkers, proxy, defaultPassword, defaultDomain)
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

func redactPasswordForDisplay(password string) string {
	if password == "" {
		return "(random)"
	}
	return "[redacted]"
}

func redactProxyForDisplay(proxy string) string {
	if proxy == "" {
		return ""
	}
	parsed, err := url.Parse(proxy)
	if err != nil {
		return "[redacted]"
	}
	if parsed.User != nil {
		parsed.User = url.UserPassword("[redacted]", "[redacted]")
	}
	return parsed.String()
}

func runMain() int {
	return executeWithIO(os.Stdin, os.Stdout, os.Stderr)
}
