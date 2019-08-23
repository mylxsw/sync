package config

type Config struct {
	FileTransferBufferSize int64
	RPCListenAddr          string
	DB                     string
	FileSyncWorkerNum      int
	JobHistoryKeepSize     int64
}
