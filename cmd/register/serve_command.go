package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"

	"github.com/monet88/chatgpt-creator/internal/email"
	"github.com/monet88/chatgpt-creator/internal/phone"
	"github.com/monet88/chatgpt-creator/internal/register"
	"github.com/monet88/chatgpt-creator/internal/web"
	"github.com/spf13/cobra"
)

func newServeCommand(out, errOut io.Writer) *cobra.Command {
	var (
		port      int
		noBrowser bool
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Launch the web UI for account registration",
		RunE: func(cmd *cobra.Command, args []string) error {
			run := func(ctx context.Context, cfg web.JobConfig, w io.Writer) (web.JobResult, error) {
				var otpProvider email.OTPProvider
				if cfg.CloudflareMailURL != "" {
					otpProvider = email.NewCloudflareTempMailProvider(cfg.CloudflareMailURL, cfg.CloudflareMailToken)
				}

				var createTempEmail func(string) (emailAddr, mailboxURL string, err error)
				if cfg.CloudflareMailURL != "" {
					createTempEmail = email.CreateCloudflareTempEmail(cfg.CloudflareMailURL, cfg.CloudflareMailToken)
				}

				var viOTPClient *phone.ViOTPClient
				if cfg.ViOTPToken != "" && cfg.ViOTPServiceID > 0 {
					viOTPClient = phone.NewViOTPClient(cfg.ViOTPToken)
				}

				providers := register.ProviderOptions{
					OTPProvider:     otpProvider,
					CreateTempEmail: createTempEmail,
				}
				if viOTPClient != nil {
					providers.PhoneProvider = viOTPClient
					providers.ViOTPServiceID = cfg.ViOTPServiceID
				}
				if cfg.CodexEnabled && cfg.CodexOutput != "" {
					providers.CodexEnabled = true
					providers.CodexOutput = cfg.CodexOutput
					providers.PanelOutputDir = cfg.PanelOutputDir
				}

				pacing := cfg.Pacing
				if pacing == "" {
					pacing = "human"
				}
				pacingProfile, err := register.ParsePacingProfile(pacing)
				if err != nil {
					return web.JobResult{}, fmt.Errorf("invalid pacing %q: %w", pacing, err)
				}

				var result register.BatchResult
				register.WithDiagnosticWriter(w, func() {
					opts := register.DefaultBatchOptionsForCLI(cfg.Total)
					opts.PacingProfile = pacingProfile
					result = register.RunBatchForCLIWithProviders(
						ctx, cfg.Total, cfg.Output, cfg.Workers,
						cfg.Proxy, cfg.Password, cfg.Domain, opts, providers,
					)
				})

				return web.JobResult{
					Success:    result.Success,
					Failed:     result.Failures,
					Target:     result.Target,
					StopReason: string(result.StopReason),
				}, nil
			}

			srv := web.NewServer(port, run)
			addr := fmt.Sprintf("http://localhost:%d", port)
			fmt.Fprintf(out, "Web UI running at %s\n", addr)
			if !noBrowser {
				openBrowser(addr)
			}
			return srv.Start(cmd.Context())
		},
	}

	cmd.Flags().IntVar(&port, "port", 8899, "Port to serve the web UI on")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Do not open the browser automatically")
	return cmd
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	_ = cmd.Start()
}
