package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dijiacoder/MetaNodeStakeSync/app/service"
	"github.com/dijiacoder/MetaNodeStakeSync/app/service/config"
	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the MetaNodeStakeSync daemon",
	Long:  `Start the MetaNodeStakeSync daemon to listen and process blockchain events.`,
	Run: func(cmd *cobra.Command, args []string) {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)

		onSyncExit := make(chan error, 1)

		go func() {
			defer wg.Done()

			cfg, err := config.UnmarshalCmdConfig()
			if err != nil {
				fmt.Println("Failed to unmarshal config", zap.Error(err))
				onSyncExit <- err
				return
			}

			logx.MustSetup(cfg.Log)
			logx.Info("sync server start", zap.Any("config", cfg))

			s, err := service.New(ctx, cfg)
			if err != nil {
				logx.Error("Failed to create sync server", zap.Error(err))
				onSyncExit <- err
				return
			}

			s.Start()

			if cfg.Monitor.PprofEnable { // 开启pprof，用于性能监控
				srv := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", cfg.Monitor.PprofPort)}
				// 启动 pprof 服务
				go func() {
					if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						onSyncExit <- err
					}
				}()
				// 在取消时优雅关闭
				go func() {
					<-ctx.Done()
					_ = srv.Shutdown(context.Background())
				}()
			}
		}()

		// 信号通知chan
		onSignal := make(chan os.Signal, 1)
		// 注册信号处理器，优雅退出
		signal.Notify(onSignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-onSignal:
			switch sig {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
				cancel()
				logx.Info("Exit by signal", zap.String("signal", sig.String()))
			}
		case err := <-onSyncExit:
			cancel()
			logx.Error("Exit by error", zap.Error(err))
		}

		wg.Wait()
	},
}

func init() {
	// 将api初始化命令添加到主命令中
	rootCmd.AddCommand(DaemonCmd)
}
