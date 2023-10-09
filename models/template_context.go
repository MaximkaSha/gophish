package models

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/url"
	"path"
	"text/template"

	"github.com/skip2/go-qrcode"
	"github.com/xyproto/png2svg"
)

// TemplateContext is an interface that allows both campaigns and email
// requests to have a PhishingTemplateContext generated for them.
type TemplateContext interface {
	getFromAddress() string
	getBaseURL() string
}

// PhishingTemplateContext is the context that is sent to any template, such
// as the email or landing page content.
type PhishingTemplateContext struct {
	From        string
	URL         string
	Tracker     string
	TrackingURL string
	RId         string
	BaseURL     string
	QRCode      string // added QR Code
	QRFile      string // QRCode base64 file
	BaseRecipient
}

// NewPhishingTemplateContext returns a populated PhishingTemplateContext,
// parsing the correct fields from the provided TemplateContext and recipient.
func NewPhishingTemplateContext(ctx TemplateContext, r BaseRecipient, rid string) (PhishingTemplateContext, error) {
	f, err := mail.ParseAddress(ctx.getFromAddress())
	if err != nil {
		return PhishingTemplateContext{}, err
	}
	fn := f.Name
	if fn == "" {
		fn = f.Address
	}
	templateURL, err := ExecuteTemplate(ctx.getBaseURL(), r)
	if err != nil {
		return PhishingTemplateContext{}, err
	}

	// For the base URL, we'll reset the the path and the query
	// This will create a URL in the form of http://example.com
	baseURL, err := url.Parse(templateURL)
	if err != nil {
		return PhishingTemplateContext{}, err
	}
	baseURL.Path = ""
	baseURL.RawQuery = ""

	phishURL, _ := url.Parse(templateURL)
	q := phishURL.Query()
	q.Set(RecipientParameter, rid)
	phishURL.RawQuery = q.Encode()
	// Generate QR code phishing link and pack it to base64
	qrCodeImageData, taskError := qrcode.Encode(phishURL.String(), qrcode.High, 256)
	if taskError != nil {
		return PhishingTemplateContext{}, err
	}
	QRCode := base64.StdEncoding.EncodeToString(qrCodeImageData)

	trackingURL, _ := url.Parse(templateURL)
	trackingURL.Path = path.Join(trackingURL.Path, "/track")
	trackingURL.RawQuery = q.Encode()

	return PhishingTemplateContext{
		BaseRecipient: r,
		BaseURL:       baseURL.String(),
		URL:           phishURL.String(),
		TrackingURL:   trackingURL.String(),
		Tracker:       "<img alt='' style='display: none' src='" + trackingURL.String() + "'/>",
		From:          fn,
		RId:           rid,
		QRCode:        "<img src='cid:" + "qr.png" + "'>", // cid for qr code
		QRFile:        QRCode,                             // base64 coded png
	}, nil
}

// ExecuteTemplate creates a templated string based on the provided
// template body and data.
func ExecuteTemplate(text string, data interface{}) (string, error) {
	buff := bytes.Buffer{}
	tmpl, err := template.New("template").Parse(text)
	if err != nil {
		return buff.String(), err
	}
	err = tmpl.Execute(&buff, data)
	return buff.String(), err
}

// ValidationContext is used for validating templates and pages
type ValidationContext struct {
	FromAddress string
	BaseURL     string
}

func (vc ValidationContext) getFromAddress() string {
	return vc.FromAddress
}

func (vc ValidationContext) getBaseURL() string {
	return vc.BaseURL
}

// ValidateTemplate ensures that the provided text in the page or template
// uses the supported template variables correctly.
func ValidateTemplate(text string) error {
	vc := ValidationContext{
		FromAddress: "foo@bar.com",
		BaseURL:     "http://example.com",
	}
	td := Result{
		BaseRecipient: BaseRecipient{
			Email:     "foo@bar.com",
			FirstName: "Foo",
			LastName:  "Bar",
			Position:  "Test",
		},
		RId: "123456",
	}
	ptx, err := NewPhishingTemplateContext(vc, td.BaseRecipient, td.RId)
	if err != nil {
		return err
	}
	_, err = ExecuteTemplate(text, ptx)
	if err != nil {
		return err
	}
	return nil
}

func PNG2SVG(filename string) error {
	fmt.Println("Converting png 2 svg")

	var (
		box          *png2svg.Box
		x, y         int
		expanded     bool
		lastx, lasty int
		lastLine     int // one message per line / y coordinate
		done         bool
	)

	img, err := png2svg.ReadPNG(filename, false)
	if err != nil {
		return err
	}

	height := img.Bounds().Max.Y - img.Bounds().Min.Y

	pi := png2svg.NewPixelImage(img, false)
	pi.SetColorOptimize(false)

	percentage := 0
	lastPercentage := 0

	if true {
		if true {
			fmt.Print("Placing rectangles... 0%")
		}

		// Cover pixels by creating expanding rectangles, as long as there are uncovered pixels
		for true && !done {

			// Select the first uncovered pixel, searching from the given coordinate
			x, y = pi.FirstUncovered(lastx, lasty)

			if true && y != lastLine {
				lastPercentage = percentage
				percentage = int((float64(y) / float64(height)) * 100.0)
				png2svg.Erase(len(fmt.Sprintf("%d%%", lastPercentage)))
				fmt.Printf("%d%%", percentage)
				lastLine = y
			}

			// Create a box at that location
			box = pi.CreateBox(x, y)
			// Expand the box to the right and downwards, until it can not expand anymore
			expanded = pi.Expand(box)

			// Use the expanded box. Color pink if it is > 1x1, and colorPink is true
			pi.CoverBox(box, expanded && false, false)

			// Check if we are done, searching from the current x,y
			done = pi.Done(x, y)
		}

		if true {
			png2svg.Erase(len(fmt.Sprintf("%d%%", lastPercentage)))
			fmt.Println("100%")
		}
	}

	return pi.WriteSVG("qr.svg")
}
