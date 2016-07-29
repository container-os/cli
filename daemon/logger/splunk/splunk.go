// Package splunk provides the log driver for forwarding server logs to
// Splunk HTTP Event Collector endpoint.
package splunk

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/daemon/logger/loggerutils"
	"github.com/docker/docker/pkg/urlutil"
)

const (
	driverName                  = "splunk"
	splunkURLKey                = "splunk-url"
	splunkTokenKey              = "splunk-token"
	splunkSourceKey             = "splunk-source"
	splunkSourceTypeKey         = "splunk-sourcetype"
	splunkIndexKey              = "splunk-index"
	splunkCAPathKey             = "splunk-capath"
	splunkCANameKey             = "splunk-caname"
	splunkInsecureSkipVerifyKey = "splunk-insecureskipverify"
	envKey                      = "env"
	labelsKey                   = "labels"
	tagKey                      = "tag"
)

type splunkLogger struct {
	client    *http.Client
	transport *http.Transport

	url         string
	auth        string
	nullMessage *splunkMessage
}

type splunkMessage struct {
	Event      splunkMessageEvent `json:"event"`
	Time       string             `json:"time"`
	Host       string             `json:"host"`
	Source     string             `json:"source,omitempty"`
	SourceType string             `json:"sourcetype,omitempty"`
	Index      string             `json:"index,omitempty"`
}

type splunkMessageEvent struct {
	Line   string            `json:"line"`
	Source string            `json:"source"`
	Tag    string            `json:"tag,omitempty"`
	Attrs  map[string]string `json:"attrs,omitempty"`
}

func init() {
	if err := logger.RegisterLogDriver(driverName, New); err != nil {
		logrus.Fatal(err)
	}
	if err := logger.RegisterLogOptValidator(driverName, ValidateLogOpt); err != nil {
		logrus.Fatal(err)
	}
}

// New creates splunk logger driver using configuration passed in context
func New(ctx logger.Context) (logger.Logger, error) {
	hostname, err := ctx.Hostname()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot access hostname to set source field", driverName)
	}

	// Parse and validate Splunk URL
	splunkURL, err := parseURL(ctx)
	if err != nil {
		return nil, err
	}

	// Splunk Token is required parameter
	splunkToken, ok := ctx.Config[splunkTokenKey]
	if !ok {
		return nil, fmt.Errorf("%s: %s is expected", driverName, splunkTokenKey)
	}

	tlsConfig := &tls.Config{}

	// Splunk is using autogenerated certificates by default,
	// allow users to trust them with skipping verification
	if insecureSkipVerifyStr, ok := ctx.Config[splunkInsecureSkipVerifyKey]; ok {
		insecureSkipVerify, err := strconv.ParseBool(insecureSkipVerifyStr)
		if err != nil {
			return nil, err
		}
		tlsConfig.InsecureSkipVerify = insecureSkipVerify
	}

	// If path to the root certificate is provided - load it
	if caPath, ok := ctx.Config[splunkCAPathKey]; ok {
		caCert, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, err
		}
		caPool := x509.NewCertPool()
		caPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caPool
	}

	if caName, ok := ctx.Config[splunkCANameKey]; ok {
		tlsConfig.ServerName = caName
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := &http.Client{
		Transport: transport,
	}

	var nullMessage = &splunkMessage{
		Host: hostname,
	}

	// Optional parameters for messages
	nullMessage.Source = ctx.Config[splunkSourceKey]
	nullMessage.SourceType = ctx.Config[splunkSourceTypeKey]
	nullMessage.Index = ctx.Config[splunkIndexKey]

	tag, err := loggerutils.ParseLogTag(ctx, loggerutils.DefaultTemplate)
	if err != nil {
		return nil, err
	}
	nullMessage.Event.Tag = tag
	nullMessage.Event.Attrs = ctx.ExtraAttributes(nil)

	logger := &splunkLogger{
		client:      client,
		transport:   transport,
		url:         splunkURL.String(),
		auth:        "Splunk " + splunkToken,
		nullMessage: nullMessage,
	}

	err = verifySplunkConnection(logger)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func (l *splunkLogger) Log(msg *logger.Message) error {
	// Construct message as a copy of nullMessage
	message := *l.nullMessage
	message.Time = fmt.Sprintf("%f", float64(msg.Timestamp.UnixNano())/1000000000)
	message.Event.Line = string(msg.Line)
	message.Event.Source = msg.Source

	jsonEvent, err := json.Marshal(&message)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", l.url, bytes.NewBuffer(jsonEvent))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", l.auth)
	res, err := l.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		var body []byte
		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s: failed to send event - %s - %s", driverName, res.Status, body)
	}
	io.Copy(ioutil.Discard, res.Body)
	return nil
}

func (l *splunkLogger) Close() error {
	l.transport.CloseIdleConnections()
	return nil
}

func (l *splunkLogger) Name() string {
	return driverName
}

// ValidateLogOpt looks for all supported by splunk driver options
func ValidateLogOpt(cfg map[string]string) error {
	for key := range cfg {
		switch key {
		case splunkURLKey:
		case splunkTokenKey:
		case splunkSourceKey:
		case splunkSourceTypeKey:
		case splunkIndexKey:
		case splunkCAPathKey:
		case splunkCANameKey:
		case splunkInsecureSkipVerifyKey:
		case envKey:
		case labelsKey:
		case tagKey:
		default:
			return fmt.Errorf("unknown log opt '%s' for %s log driver", key, driverName)
		}
	}
	return nil
}

func parseURL(ctx logger.Context) (*url.URL, error) {
	splunkURLStr, ok := ctx.Config[splunkURLKey]
	if !ok {
		return nil, fmt.Errorf("%s: %s is expected", driverName, splunkURLKey)
	}

	splunkURL, err := url.Parse(splunkURLStr)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse %s as url value in %s", driverName, splunkURLStr, splunkURLKey)
	}

	if !urlutil.IsURL(splunkURLStr) ||
		!splunkURL.IsAbs() ||
		(splunkURL.Path != "" && splunkURL.Path != "/") ||
		splunkURL.RawQuery != "" ||
		splunkURL.Fragment != "" {
		return nil, fmt.Errorf("%s: expected format scheme://dns_name_or_ip:port for %s", driverName, splunkURLKey)
	}

	splunkURL.Path = "/services/collector/event/1.0"

	return splunkURL, nil
}

func verifySplunkConnection(l *splunkLogger) error {
	req, err := http.NewRequest("OPTIONS", l.url, nil)
	if err != nil {
		return err
	}
	res, err := l.client.Do(req)
	if err != nil {
		return err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if res.StatusCode != http.StatusOK {
		var body []byte
		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s: failed to verify connection - %s - %s", driverName, res.Status, body)
	}
	return nil
}
