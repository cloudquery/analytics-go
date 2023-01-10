package analytics

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/backo-go"
)

// Instances of this type carry the different configuration options that may
// be set when instantiating a client.
//
// Each field's zero-value is either meaningful or interpreted as using the
// default value defined by the library.
type Config struct {

	// Deprecated: Endpoint is deprecated, will be removed in next releases. Use DataPlaneUrl.
	Endpoint string

	// The endpoint to which the client connect and send their messages, set to
	// `DefaultEndpoint` by default.
	DataPlaneUrl string

	// The flushing interval of the client. Messages will be sent when they've
	// been queued up to the maximum batch size or when the flushing interval
	// timer triggers.
	Interval time.Duration

	// The HTTP transport used by the client, this allows an application to
	// redefine how requests are being sent at the HTTP level (for example,
	// to change the connection pooling policy).
	// If none is specified the client uses `http.DefaultTransport`.
	Transport http.RoundTripper

	// The logger used by the client to output info or error messages when that
	// are generated by background operations.
	// If none is specified the client uses a standard logger that outputs to
	// `os.Stderr`.
	Logger Logger

	// The callback object that will be used by the client to notify the
	// application when messages sends to the backend API succeeded or failed.
	Callback Callback

	// The maximum number of messages that will be sent in one API call.
	// Messages will be sent when they've been queued up to the maximum batch
	// size or when the flushing interval timer triggers.
	// Note that the API will still enforce a 500KB limit on each HTTP request
	// which is independent from the number of embedded messages.
	BatchSize int

	// When set to true the client will send more frequent and detailed messages
	// to its logger.
	Verbose bool

	// The default context set on each message sent by the client.
	DefaultContext *Context

	// The retry policy used by the client to resend requests that have failed.
	// The function is called with how many times the operation has been retried
	// and is expected to return how long the client should wait before trying
	// again.
	// If not set the client will fallback to use a default retry policy.
	RetryAfter func(int) time.Duration

	// A function called by the client to generate unique message identifiers.
	// The client uses a UUID generator if none is provided.
	// This field is not exported and only exposed internally to let unit tests
	// mock the id generation.
	uid func() string

	// A function called by the client to get the current time, `time.Now` is
	// used by default.
	// This field is not exported and only exposed internally to let unit tests
	// mock the current time.
	now func() time.Time

	// The maximum number of goroutines that will be spawned by a client to send
	// requests to the backend API.
	// This field is not exported and only exposed internally to let unit tests
	// mock the current time.
	maxConcurrentRequests int

	//This variable will disable checking for the cluster-info end point and
	//split the payload at node level for multi node setup
	NoProxySupport bool

	// Maximum bytes in a message
	MaxMessageBytes int

	// Maximum bytes in a batch
	MaxBatchBytes int

	// Deprecated: Gzip is deprecated, will be removed in next releases. Use DisableGzip.
	Gzip int

	// Disable/enable gzip support.
	DisableGzip bool
}

// This constant sets the default endpoint to which client instances send
// messages if none was explictly set.

const DefaultEndpoint = "https://hosted.rudderlabs.com"

// This constant sets the default flush interval used by client instances if
// none was explicitly set.
const DefaultInterval = 5 * time.Second

// This constant sets the default batch size used by client instances if none
// was explicitly set.
const DefaultBatchSize = 250

// Verifies that fields that don't have zero-values are set to valid values,
// returns an error describing the problem if a field was invalid.
func (c *Config) validate() error {
	if c.Interval < 0 {
		return ConfigError{
			Reason: "negative time intervals are not supported",
			Field:  "Interval",
			Value:  c.Interval,
		}
	}

	if c.BatchSize < 0 {
		return ConfigError{
			Reason: "negative batch sizes are not supported",
			Field:  "BatchSize",
			Value:  c.BatchSize,
		}
	}

	if c.MaxMessageBytes < 0 {
		return ConfigError{
			Reason: "negetive value is not supported for MaxMessageBytes",
			Field:  "MaxMessageBytes",
			Value:  c.MaxMessageBytes,
		}
	}

	if c.MaxBatchBytes < 0 {
		return ConfigError{
			Reason: "negetive value is not supported for MaxBatchBytes",
			Field:  "MaxBatchBytes",
			Value:  c.MaxBatchBytes,
		}
	}

	return nil
}

// Given a config object as argument the function will set all zero-values to
// their defaults and return the modified object.
func makeConfig(c Config) Config {
	if len(c.DataPlaneUrl) > 0 {
		c.Endpoint = c.DataPlaneUrl
	}

	if len(c.Endpoint) == 0 && len(c.DataPlaneUrl) == 0 {
		c.Endpoint = DefaultEndpoint
	}

	if c.Interval == 0 {
		c.Interval = DefaultInterval
	}

	if c.Transport == nil {
		c.Transport = http.DefaultTransport
	}

	if c.Logger == nil {
		c.Logger = newDefaultLogger()
	}

	if c.BatchSize == 0 {
		c.BatchSize = DefaultBatchSize
	}

	if c.DefaultContext == nil {
		c.DefaultContext = &Context{}
	}

	if c.RetryAfter == nil {
		c.RetryAfter = backo.NewBacko(time.Millisecond*100, 2, 1, time.Second*30).Duration
	}

	if c.uid == nil {
		c.uid = uid
	}

	if c.now == nil {
		c.now = time.Now
	}

	if c.maxConcurrentRequests == 0 {
		c.maxConcurrentRequests = 1000
	}

	if c.MaxMessageBytes == 0 {
		c.MaxMessageBytes = defMaxMessageBytes
	}

	if c.MaxBatchBytes == 0 {
		c.MaxBatchBytes = defMaxBatchBytes
	}

	if c.Gzip != 0 {
		c.DisableGzip = true
	}

	// We always overwrite the 'library' field of the default context set on the
	// client because we want this information to be accurate.
	c.DefaultContext.Library = LibraryInfo{
		Name:    "analytics-go",
		Version: Version,
	}
	return c
}

// This function returns a string representation of a UUID, it's the default
// function used for generating unique IDs.
func uid() string {
	return uuid.NewString()
}
