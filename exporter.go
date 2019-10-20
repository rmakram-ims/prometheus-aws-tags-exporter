package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	flagInterval = "interval"

	lambdaName = ""
)

var (
	gauge *prometheus.GaugeVec
)

type ExporterCmd struct {
	sess *session.Session
}

func (e *ExporterCmd) contains(a []string, x string) bool {
	for _, n := range a {
		if strings.Compare(e.formatMetricName(x), n) == 0 {
			return true
		}
	}
	return false
}

func (e *ExporterCmd) formatMetricName(a string) string {
	return strings.Replace(strings.Replace(a, "-", "_", -1), ":", "_", -1)
}

// session() only creates one sessions
func (e *ExporterCmd) session() *session.Session {
	if e.sess == nil {
		sess := session.Must(session.NewSession())
		e.sess = sess
	}
	return e.sess
}

func (e *ExporterCmd) getSupportedTags() []string {
	return []string{
		e.formatMetricName("Version"),
	}
}

func (e *ExporterCmd) getLambdaTags(name string) (map[string]string, error) {
	svc := lambda.New(e.session())
	out, err := svc.GetFunction(&lambda.GetFunctionInput{
		FunctionName: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	r := make(map[string]string)

	s := e.getSupportedTags()
	for _, i := range s {
		r[i] = ""
	}

	for k, v := range out.Tags {
		f := e.formatMetricName(k)
		if e.contains(s, f) {
			r[f] = aws.StringValue(v)
		}
	}

	return r, nil
}

func (e *ExporterCmd) updateLambdaTags() {
	tags, err := e.getLambdaTags(lambdaName)
	if err == nil && len(tags) > 0 {
		gauge.With(tags).Add(1)
	}
}

func (e *ExporterCmd) execute(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	intervalStr, err := flags.GetString(flagInterval)
	if err != nil {
		return err
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return err
	}

	gauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: fmt.Sprintf("%s_metric", e.formatMetricName(lambdaName)),
		Help: fmt.Sprintf("Shows all the tags for Lambda %s", lambdaName),
	}, e.getSupportedTags())
	prometheus.MustRegister(gauge)

	go func() {
		timerCh := time.Tick(time.Duration(interval) * time.Second)
		for range timerCh {
			e.updateLambdaTags()
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to serve on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
	return nil
}

// NewExporterCmd defines all flags for the export command
func NewExporterCmd(e *ExporterCmd) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [options]",
		Short: "Export AWS resource tags as metrics",
		Long: fmt.Sprintf(`Collect and export AWS resource Tags as Prometheus metrics
		
Example usage:
%s export `, appName),
		RunE: e.execute,
	}

	cmd.Flags().String(flagInterval, "10", "interval of metric updates (seconds)")

	return cmd
}
