package router

// Logger interface represents error logger.
//
// This package needs to log error on ServeHTTP. So
// this interface is used for that.
type Logger interface {
	Printf(format string, v ...interface{})
}
