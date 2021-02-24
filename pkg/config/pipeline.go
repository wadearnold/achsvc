// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"fmt"
	"os"
	"text/template"

	"github.com/moov-io/paygate/pkg/util"
)

var (
	DefaultEmailTemplate = template.Must(template.New("email").Parse(`
A file has been {{ .Verb }}ed{{ if .Hostname }}{{ if eq .Verb "upload" }} to{{ else }} from{{end}} {{ .Hostname }}{{end}} - {{ .Filename }}
Name: {{ .CompanyName }}
Debits:  ${{ .DebitTotal | printf "%.2f" }}
Credits: ${{ .CreditTotal | printf "%.2f" }}

Batches: {{ .BatchCount }}
Total Entries: {{ .EntryCount }}
`))
)

type Pipeline struct {
	PreUpload     *PreUpload
	Output        *Output
	Merging       *Merging
	AuditTrail    *AuditTrail
	Stream        *StreamPipeline
	Notifications *PipelineNotifications
}

func (cfg Pipeline) Validate() error {
	if err := cfg.PreUpload.Validate(); err != nil {
		return fmt.Errorf("pre-upload: %v", err)
	}
	if err := cfg.Output.Validate(); err != nil {
		return fmt.Errorf("output: %v", err)
	}
	if err := cfg.Merging.Validate(); err != nil {
		return fmt.Errorf("merging: %v", err)
	}
	if err := cfg.AuditTrail.Validate(); err != nil {
		return fmt.Errorf("audit-trail: %v", err)
	}
	if err := cfg.Stream.Validate(); err != nil {
		return fmt.Errorf("stream: %v", err)
	}
	if err := cfg.Notifications.Validate(); err != nil {
		return fmt.Errorf("notifications: %v", err)
	}
	return nil
}

type PreUpload struct {
	GPG *GPG
}

func (cfg *PreUpload) Validate() error {
	if cfg == nil {
		return nil
	}
	if cfg.GPG != nil && cfg.GPG.KeyFile == "" {
		return errors.New("gpg: missing key file")
	}
	return nil
}

type GPG struct {
	KeyFile string
	Signer  *Signer
}

type Signer struct {
	KeyFile     string
	KeyPassword string
}

func (cfg *Signer) Password() string {
	return util.Or(os.Getenv("PIPELINE_SIGNING_KEY_PASSWORD"), cfg.KeyPassword)
}

type Output struct {
	Format string
}

func (cfg *Output) Validate() error {
	return nil
}

type Merging struct {
	Directory      string
	FlattenBatches *FlattenBatches
}

func (cfg *Merging) Validate() error {
	return nil
}

type FlattenBatches struct{}

type AuditTrail struct {
	BucketURI string
	GPG       *GPG
}

func (cfg *AuditTrail) Validate() error {
	if cfg == nil {
		return nil
	}
	if cfg.BucketURI == "" {
		return errors.New("missing bucket_uri")
	}
	return nil
}

type StreamPipeline struct {
	InMem *InMemPipeline
	Kafka *KafkaPipeline
}

func (cfg *StreamPipeline) Validate() error {
	if cfg == nil {
		return nil
	}
	if cfg.InMem != nil && cfg.InMem.URL == "" {
		return errors.New("inmem: missing stream url")
	}
	if k := cfg.Kafka; k != nil {
		if len(k.Brokers) == 0 || k.Group == "" || k.Topic == "" {
			return errors.New("kafka: missing brokers, group, or topic")
		}
	}
	return nil
}

type InMemPipeline struct {
	URL string
}

type KafkaPipeline struct {
	Brokers []string
	Group   string
	Topic   string
}

type PipelineNotifications struct {
	Email     *Email
	PagerDuty *PagerDuty
	Slack     *Slack
}

func (cfg *PipelineNotifications) Validate() error {
	if cfg == nil {
		return nil
	}
	if e := cfg.Email; e != nil {
		if e.From == "" || len(e.To) == 0 || e.ConnectionURI == "" || e.CompanyName == "" {
			return errors.New("email: missing configs")
		}
	}
	if cfg.PagerDuty != nil && cfg.PagerDuty.ApiKey == "" {
		return errors.New("pagerduty: missing api key")
	}
	if err := cfg.Slack.Validate(); err != nil {
		return err
	}
	return nil
}

type Email struct {
	From string
	To   []string

	// ConnectionURI is a URI used to connect with a remote SFTP server.
	// This config typically needs to contain enough values to successfully
	// authenticate with the server.
	// - insecure_skip_verify is an optional parameter for disabling certificate verification
	//
	// Example: smtps://user:pass@localhost:1025/?insecure_skip_verify=true
	ConnectionURI string

	Template    string
	CompanyName string
}

func (e *Email) Tmpl() *template.Template {
	if e == nil || e.Template == "" {
		return DefaultEmailTemplate
	}
	return template.Must(template.New("custom-email").Parse(e.Template))
}

type PagerDuty struct {
	ApiKey     string
	From       string
	ServiceKey string
}

type Slack struct {
	WebhookURL string
}

func (cfg *Slack) Validate() error {
	if cfg == nil {
		return nil
	}
	if cfg.WebhookURL == "" {
		return errors.New("slack: missing webhook url")
	}
	return nil
}
