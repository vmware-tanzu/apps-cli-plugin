//   Copyright 2016 Wercker Holding BV
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package stern

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/fatih/color"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

// RFC3339Nano with trailing zeros
const TimestampFormatDefault = "2006-01-02T15:04:05.000000000Z07:00"

// time.DateTime without year
const TimestampFormatShort = "01-02 15:04:05"

type Tail struct {
	clientset corev1client.CoreV1Interface

	NodeName       string
	Namespace      string
	PodName        string
	ContainerName  string
	Options        *TailOptions
	closed         chan struct{}
	podColor       *color.Color
	containerColor *color.Color
	tmpl           *template.Template
	last           struct {
		timestamp string // RFC3339 timestamp (not RFC3339Nano)
		lines     int    // the number of lines seen during this timestamp
	}
	resumeRequest *ResumeRequest
	out           io.Writer
	errOut        io.Writer
}

type TailOptions struct {
	Timestamps      bool
	TimestampFormat string
	Location        *time.Location

	SinceSeconds *int64
	SinceTime    *metav1.Time
	Exclude      []*regexp.Regexp
	Include      []*regexp.Regexp
	Highlight    []*regexp.Regexp
	Namespace    bool
	TailLines    *int64
	Follow       bool
	OnlyLogLines bool

	// regexp for highlighting the matched string
	reHightlight *regexp.Regexp
}

type ResumeRequest struct {
	Timestamp   string // RFC3339 timestamp (not RFC3339Nano)
	LinesToSkip int    // the number of lines to skip during this timestamp
}

func (o TailOptions) IsExclude(msg string) bool {
	for _, rex := range o.Exclude {
		if rex.MatchString(msg) {
			return true
		}
	}

	return false
}

func (o TailOptions) IsInclude(msg string) bool {
	if len(o.Include) == 0 {
		return true
	}

	for _, rin := range o.Include {
		if rin.MatchString(msg) {
			return true
		}
	}

	return false
}

var colorHighlight = color.New(color.FgRed, color.Bold).SprintFunc()

func (o TailOptions) HighlightMatchedString(msg string) string {
	highlight := append(o.Include, o.Highlight...)
	if len(highlight) == 0 {
		return msg
	}

	if o.reHightlight == nil {
		ss := make([]string, len(highlight))
		for i, hl := range highlight {
			ss[i] = hl.String()
		}

		// We expect a longer match
		sort.Slice(ss, func(i, j int) bool {
			return len(ss[i]) > len(ss[j])
		})

		o.reHightlight = regexp.MustCompile("(" + strings.Join(ss, "|") + ")")
	}

	msg = o.reHightlight.ReplaceAllStringFunc(msg, func(part string) string {
		return colorHighlight(part)
	})

	return msg
}

func (o TailOptions) UpdateTimezoneAndFormat(timestamp string) (string, error) {
	t, err := time.ParseInLocation(time.RFC3339Nano, timestamp, time.UTC)
	if err != nil {
		return "", errors.New("missing timestamp")
	}
	format := TimestampFormatDefault
	if o.TimestampFormat != "" {
		format = o.TimestampFormat
	}
	return t.In(o.Location).Format(format), nil
}

// NewTail returns a new tail for a Kubernetes container inside a pod
func NewTail(clientset corev1client.CoreV1Interface, nodeName, namespace, podName, containerName string, tmpl *template.Template, out, errOut io.Writer, options *TailOptions) *Tail {
	podColor, containerColor := determineColor(podName)

	return &Tail{
		clientset:      clientset,
		NodeName:       nodeName,
		Namespace:      namespace,
		PodName:        podName,
		ContainerName:  containerName,
		Options:        options,
		closed:         make(chan struct{}),
		tmpl:           tmpl,
		podColor:       podColor,
		containerColor: containerColor,

		out:    out,
		errOut: errOut,
	}
}

var colorList = [][2]*color.Color{
	{color.New(color.FgHiCyan), color.New(color.FgCyan)},
	{color.New(color.FgHiGreen), color.New(color.FgGreen)},
	{color.New(color.FgHiMagenta), color.New(color.FgMagenta)},
	{color.New(color.FgHiYellow), color.New(color.FgYellow)},
	{color.New(color.FgHiBlue), color.New(color.FgBlue)},
	{color.New(color.FgHiRed), color.New(color.FgRed)},
}

func determineColor(podName string) (podColor, containerColor *color.Color) {
	hash := fnv.New32()
	_, _ = hash.Write([]byte(podName))
	idx := hash.Sum32() % uint32(len(colorList))

	colors := colorList[idx]
	return colors[0], colors[1]
}

// Start starts tailing
func (t *Tail) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-t.closed
		cancel()
	}()

	t.printStarting()

	req := t.clientset.Pods(t.Namespace).GetLogs(t.PodName, &corev1.PodLogOptions{
		Follow:       t.Options.Follow,
		Timestamps:   true,
		Container:    t.ContainerName,
		SinceSeconds: t.Options.SinceSeconds,
		SinceTime:    t.Options.SinceTime,
		TailLines:    t.Options.TailLines,
	})

	err := t.ConsumeRequest(ctx, req)

	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}

func (t *Tail) Resume(ctx context.Context, resumeRequest *ResumeRequest) error {
	sinceTime, err := resumeRequest.sinceTime()
	if err != nil {
		fmt.Fprintf(t.errOut, "failed to resume: %s, fallback to Start()\n", err)
		return t.Start(ctx)
	}
	t.resumeRequest = resumeRequest
	t.Options.SinceTime = sinceTime
	t.Options.SinceSeconds = nil
	t.Options.TailLines = nil
	return t.Start(ctx)
}

// Close stops tailing
func (t *Tail) Close() {
	t.printStopping()

	close(t.closed)
}

func (t *Tail) printStarting() {
	if !t.Options.OnlyLogLines {
		g := color.New(color.FgHiGreen, color.Bold).SprintFunc()
		p := t.podColor.SprintFunc()
		c := t.containerColor.SprintFunc()
		if t.Options.Namespace {
			fmt.Fprintf(t.errOut, "%s %s %s › %s\n", g("+"), p(t.Namespace), p(t.PodName), c(t.ContainerName))
		} else {
			fmt.Fprintf(t.errOut, "%s %s › %s\n", g("+"), p(t.PodName), c(t.ContainerName))
		}
	}
}

func (t *Tail) printStopping() {
	if !t.Options.OnlyLogLines {
		r := color.New(color.FgHiRed, color.Bold).SprintFunc()
		p := t.podColor.SprintFunc()
		c := t.containerColor.SprintFunc()
		if t.Options.Namespace {
			fmt.Fprintf(t.errOut, "%s %s %s › %s\n", r("-"), p(t.Namespace), p(t.PodName), c(t.ContainerName))
		} else {
			fmt.Fprintf(t.errOut, "%s %s › %s\n", r("-"), p(t.PodName), c(t.ContainerName))
		}
	}
}

// ConsumeRequest reads the data from request and writes into the out
// writer.
func (t *Tail) ConsumeRequest(ctx context.Context, request rest.ResponseWrapper) error {
	stream, err := request.Stream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	r := bufio.NewReader(stream)
	for {
		line, err := r.ReadBytes('\n')
		if len(line) != 0 {
			t.consumeLine(strings.TrimSuffix(string(line), "\n"))
		}

		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
	}
}

// Print prints a color coded log message with the pod and container names
func (t *Tail) Print(msg string) {
	vm := Log{
		Message:        msg,
		NodeName:       t.NodeName,
		Namespace:      t.Namespace,
		PodName:        t.PodName,
		ContainerName:  t.ContainerName,
		PodColor:       t.podColor,
		ContainerColor: t.containerColor,
	}

	var buf bytes.Buffer
	if err := t.tmpl.Execute(&buf, vm); err != nil {
		fmt.Fprintf(t.errOut, "expanding template failed: %s\n", err)
		return
	}

	fmt.Fprint(t.out, buf.String())
}

func (t *Tail) GetResumeRequest() *ResumeRequest {
	if t.last.timestamp == "" {
		return nil
	}
	return &ResumeRequest{Timestamp: t.last.timestamp, LinesToSkip: t.last.lines}
}

func (t *Tail) consumeLine(line string) {
	rfc3339Nano, content, err := splitLogLine(line)
	if err != nil {
		t.Print(fmt.Sprintf("[%v] %s", err, line))
		return
	}

	// PodLogOptions.SinceTime is RFC3339, not RFC3339Nano.
	// We convert it to RFC3339 to skip the lines seen during this timestamp when resuming.
	rfc3339 := removeSubsecond(rfc3339Nano)
	t.rememberLastTimestamp(rfc3339)
	if t.resumeRequest.shouldSkip(rfc3339) {
		return
	}

	if t.Options.IsExclude(content) || !t.Options.IsInclude(content) {
		return
	}

	msg := t.Options.HighlightMatchedString(content)

	if t.Options.Timestamps {
		updatedTs, err := t.Options.UpdateTimezoneAndFormat(rfc3339Nano)
		if err != nil {
			t.Print(fmt.Sprintf("[%v] %s", err, line))
			return
		}
		msg = updatedTs + " " + msg
	}

	t.Print(msg)
}

func (t *Tail) rememberLastTimestamp(timestamp string) {
	if t.last.timestamp == timestamp {
		t.last.lines++
		return
	}
	t.last.timestamp = timestamp
	t.last.lines = 1
}

// Log is the object which will be used together with the template to generate
// the output.
type Log struct {
	// Message is the log message itself
	Message string `json:"message"`

	// Node name of the pod
	NodeName string `json:"nodeName"`

	// Namespace of the pod
	Namespace string `json:"namespace"`

	// PodName of the pod
	PodName string `json:"podName"`

	// ContainerName of the container
	ContainerName string `json:"containerName"`

	PodColor       *color.Color `json:"-"`
	ContainerColor *color.Color `json:"-"`
}

func (r *ResumeRequest) sinceTime() (*metav1.Time, error) {
	sinceTime, err := time.Parse(time.RFC3339, r.Timestamp)

	if err != nil {
		return nil, err
	}
	metaTime := metav1.NewTime(sinceTime)
	return &metaTime, nil
}

func (r *ResumeRequest) shouldSkip(timestamp string) bool {
	if r == nil {
		return false
	}
	if r.Timestamp == "" {
		return false
	}
	if r.Timestamp != timestamp {
		return false
	}
	if r.LinesToSkip <= 0 {
		return false
	}
	r.LinesToSkip--
	return true
}

func splitLogLine(line string) (timestamp string, content string, err error) {
	idx := strings.IndexRune(line, ' ')
	if idx == -1 {
		return "", "", errors.New("missing timestamp")
	}
	return line[:idx], line[idx+1:], nil
}

// removeSubsecond removes the subsecond of the timestamp.
// It converts RFC3339Nano to RFC3339 fast.
func removeSubsecond(timestamp string) string {
	dot := strings.IndexRune(timestamp, '.')
	if dot == -1 {
		return timestamp
	}
	var last int
	for i := dot; i < len(timestamp); i++ {
		if unicode.IsDigit(rune(timestamp[i])) {
			last = i
		}
	}
	if last == 0 {
		return timestamp
	}
	return timestamp[:dot] + timestamp[last+1:]
}
