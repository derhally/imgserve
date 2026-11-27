package internal

// ServiceSettings holds configuration parameters for the service
type ServiceSettings struct {
	ImageDirs   []string // -imageDir values: "dir" or "name=dir", one per served directory
	CertFile    string   // Path to SSL/TLS certificate file
	CertKeyFile string   // Path to SSL/TLS certificate private key file
	LogLevel    string   // Logging verbosity level (e.g. debug, info, warn, error)
	Port        string   // Port to listen on
	CacheDir    string   // Path to the directory where temporary files are stored
}
