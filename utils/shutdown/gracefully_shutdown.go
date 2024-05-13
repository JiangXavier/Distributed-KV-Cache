package shutdown

import (
	"context"
	"leicache/utils/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func GracefullyShutdown(server *http.Server) {
	done := make(chan os.Signal, 1)

	/**
	os.Interrupt           -> ctrl+c
	syscall.SIGINT|SIGTERM -> the signal passed to the process when killing the process. such as kill -9
	*/
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done

	logger.LogrusObj.Println("closing http server gracefully...")

	if err := server.Shutdown(context.Background()); err != nil {
		logger.LogrusObj.Fatalln("closing http server gracefully failed: ", err)
	}
}
