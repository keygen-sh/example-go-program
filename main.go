package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"time"

	ui "github.com/gizak/termui"
	"github.com/mholt/archiver"
	"github.com/mitchellh/ioprogress"
)

const (
	account string = "1fddcec8-8dd3-4d8d-9b16-215cac0f9b52"
	product string = "2d130468-27aa-4c41-b064-18fc6b3046d9"
	version string = "1.0.0"
)

var (
	components map[string]Component
	license    string
)

func main() {
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	showCurrentVersion()
	showControls()

	p := ui.NewPar("")
	p.BorderLabel = "Enter your license key"
	p.TextFgColor = ui.ColorWhite
	p.BorderLabelFg = ui.ColorYellow
	p.BorderFg = ui.ColorYellow
	p.Height = 3
	p.Width = 50
	p.Y = 11
	addComponent("license-prompt", p, ComponentOpts{Order: 3})

	ui.Handle("/sys/kbd/C-c", func(ui.Event) {
		quit()
	})

	ui.Handle("/sys/kbd/<enter>", func(e ui.Event) {
		license = p.Text

		if ok := validateLicenseKey(p.Text); ok {
			p.BorderLabelFg = ui.ColorGreen
			p.BorderFg = ui.ColorGreen

			removeComponent("license-prompt")

			go run()
		} else {
			p.BorderLabelFg = ui.ColorRed
			p.BorderFg = ui.ColorRed
		}
		p.Text = ""

		updateComponent("license-prompt", p, ComponentOpts{Order: 3})
	})

	// FIXME(ezekg) termui seems to emit C-8 instead of <backspace>
	ui.Handle("/sys/kbd/C-8", func(e ui.Event) {
		if l := len(p.Text); l > 0 {
			p.Text = p.Text[0 : l-1]
		}

		updateComponent("license-prompt", p, ComponentOpts{Order: 3})
	})

	ui.Handle("/sys/kbd/<space>", func(e ui.Event) {
		p.Text += " "

		updateComponent("license-prompt", p, ComponentOpts{Order: 3})
	})

	ui.Handle("/sys/kbd", func(e ui.Event) {
		kbd := e.Data.(ui.EvtKbd)
		p.Text += kbd.KeyStr

		updateComponent("license-prompt", p, ComponentOpts{Order: 3})
	})

	render()

	ui.Loop()
}

type ComponentOpts struct {
	Order int
}

type Component struct {
	Name     string
	Bufferer ui.Bufferer
	Opts     ComponentOpts
}

type ComponentSorter []Component

func (s ComponentSorter) Len() int           { return len(s) }
func (s ComponentSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ComponentSorter) Less(i, j int) bool { return s[i].Opts.Order < s[j].Opts.Order }

func addComponent(name string, component ui.Bufferer, opts ComponentOpts) {
	if components == nil {
		components = make(map[string]Component)
	}

	components[name] = Component{name, component, opts}

	render()
}

func updateComponent(name string, component ui.Bufferer, opts ComponentOpts) {
	if _, ok := components[name]; !ok {
		return
	}
	components[name] = Component{name, component, opts}

	ui.Render(component) // Render single component update
}

func removeComponent(name string) {
	delete(components, name)

	render()
}

func render() {
	var comps []Component
	var bufs []ui.Bufferer

	// Sort components by render order
	for _, comp := range components {
		comps = append(comps, comp)
	}
	sort.Sort(ComponentSorter(comps))

	// Get bufferers to render
	for _, comp := range comps {
		bufs = append(bufs, comp.Bufferer)
	}

	ui.Render(bufs...)
}

func run() {
	g := ui.NewGauge()
	g.Percent = 0
	g.Width = 50
	g.Height = 3
	g.Y = 11
	g.BorderLabel = "Loading..."
	g.BorderLabelFg = ui.ColorGreen
	g.BorderFg = ui.ColorGreen
	g.BarColor = ui.ColorGreen
	addComponent("loading-app", g, ComponentOpts{Order: 2})

	for i := 1; i <= 100; i++ {
		time.Sleep(50 * time.Millisecond)
		g.Percent = i
		updateComponent("loading-app", g, ComponentOpts{Order: 2})
	}
	removeComponent("loading-app")

	m := ui.NewPar("")
	m.Height = 30
	m.Width = 80
	m.Text = "@@@"
	m.Border = false
	m.Y = 2
	addComponent("app", m, ComponentOpts{Order: 10})

	go enableAutoUpdates()
}

func quit() {
	ui.StopLoop()
}

type ValidationRequest struct {
	Meta ValidationMeta `json:"meta"`
}

type ValidationMeta struct {
	Key   string          `json:"key"`
	Scope ValidationScope `json:"scope,omitempty"`
}

type ValidationScope struct {
	Product string `json:"product"`
}

type Validation struct {
	Meta struct {
		Valid  bool   `json:"valid"`
		Detail string `json:"detail"`
	} `json:"meta"`
}

func validateLicenseKey(key string) bool {
	b, err := json.Marshal(ValidationRequest{ValidationMeta{key, ValidationScope{product}}})
	if err != nil {
		return false
	}

	buffer := bytes.NewBuffer(b)
	res, err := http.Post(
		fmt.Sprintf("https://api.keygen.sh/v1/accounts/%s/licenses/actions/validate-key", account),
		"application/vnd.api+json",
		buffer,
	)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return false
	}

	var v Validation
	json.NewDecoder(res.Body).Decode(&v)

	return v.Meta.Valid
}

type Update struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func enableAutoUpdates() {
	for {
		if update, ok := checkForUpdate(version, license); ok {
			showUpdateAvailable(update)

			ui.Handle("/sys/kbd/C-u", func(ui.Event) {
				if f, ok := downloadUpdate(update); ok {
					if ok := installUpdate(f); ok {
						quit() // All good. Exit so that they can restart!
					} else {
						showUpdateError(update, fmt.Sprintf("There was an error installing %s", update.Name))
					}
				} else {
					showUpdateError(update, fmt.Sprintf("There was an error downloading %s", update.Name))
				}
			})
		}
		time.Sleep(15 * time.Minute)
	}
}

func checkForUpdate(version string, license string) (*Update, bool) {
	// The default http.Get() method automatically follows redirects, so we're
	// using a client directly so we don't automatically download updates.
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://dist.keygen.sh/v1/%s/%s/update/%s-%s/zip/%s?key=%s", account, product, runtime.GOOS, runtime.GOARCH, version, license), nil)
	if err != nil {
		return nil, false
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, false
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusTemporaryRedirect {
		return nil, false
	}

	var u *Update
	json.NewDecoder(res.Body).Decode(&u)

	return u, true
}

func showControls() {
	q := ui.NewPar("Press ctrl+c to quit")
	q.TextFgColor = ui.ColorWhite
	q.BorderFg = ui.ColorCyan
	q.Border = false
	q.Height = 3
	q.Width = 80
	q.Y = 1
	addComponent("quit-control", q, ComponentOpts{Order: 8})
}

func showCurrentVersion() {
	u := ui.NewPar(fmt.Sprintf("Version %s", version))
	u.TextFgColor = ui.ColorGreen
	u.BorderFg = ui.ColorGreen
	u.Border = false
	u.Height = 3
	u.Width = 80
	addComponent("current-version", u, ComponentOpts{Order: 6})
}

func showUpdateAvailable(update *Update) {
	removeComponent("current-version")
	u := ui.NewPar(fmt.Sprintf("Press ctrl+u to update (%s)", update.Name))
	u.TextFgColor = ui.ColorGreen
	u.BorderFg = ui.ColorGreen
	u.Border = false
	u.Height = 3
	u.Width = 80
	addComponent("update-tip", u, ComponentOpts{Order: 7})
}

func showUpdateError(update *Update, msg string) {
	u := ui.NewPar(msg)
	u.TextFgColor = ui.ColorRed
	u.BorderFg = ui.ColorRed
	u.Width = 50
	u.Height = 3
	u.Y = 11
	addComponent("update-error", u, ComponentOpts{Order: 15})
}

func downloadUpdate(update *Update) (*os.File, bool) {
	res, err := http.Get(update.URL)
	if err != nil {
		return nil, false
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, false
	}

	checksum, err := base64.StdEncoding.DecodeString(res.Header.Get("Content-MD5"))
	if err != nil {
		return nil, false
	}

	_, params, err := mime.ParseMediaType(res.Header.Get("Content-Disposition"))
	if err != nil {
		return nil, false
	}

	tmp, err := ioutil.TempFile("", params["filename"])
	if err != nil {
		return nil, false
	}
	defer tmp.Close()

	// Create the progress bar
	g := ui.NewGauge()
	g.Percent = 0
	g.Width = 50
	g.Height = 3
	g.Y = 11
	g.BorderLabel = "Downloading update..."
	g.BorderLabelFg = ui.ColorGreen
	g.BorderFg = ui.ColorGreen
	g.BarColor = ui.ColorGreen
	addComponent("download-update", g, ComponentOpts{Order: 2})
	defer removeComponent("download-update")

	// Create the progress reader
	size, _ := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
	prog := &ioprogress.Reader{
		Reader:       res.Body,
		Size:         size,
		DrawInterval: time.Duration(50 * time.Millisecond),
		DrawFunc: func(size, total int64) error {
			g.Percent = int(float64(size) / float64(total) * 100)
			updateComponent("download-update", g, ComponentOpts{Order: 2})
			return nil
		},
	}

	_, err = io.Copy(tmp, prog)
	if err != nil {
		return nil, false
	}

	f, err := os.Open(tmp.Name())
	if err != nil {
		return nil, false
	}
	defer f.Close()

	// Validate checksum
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, false
	}

	md5sum := hex.EncodeToString(h.Sum(nil))
	if md5sum != string(checksum) {
		return nil, false
	}

	return f, true
}

func installUpdate(update *os.File) bool {
	ex, err := os.Executable()
	if err != nil {
		return false
	}
	tmp := fmt.Sprintf("%s-%d", ex, time.Now().UnixNano())

	// Rename current executable so that we can replace it
	err = os.Rename(ex, tmp)
	if err != nil {
		return false
	}

	// Extract zip archive and replace current executable with archived one
	err = archiver.Zip.Open(update.Name(), filepath.Dir(ex))
	if err != nil {
		os.Rename(tmp, ex) // Restore ex if extraction fails
		return false
	}

	// Get permissions for old executable so we can mirror them
	fi, err := os.Stat(tmp)
	if err != nil {
		return false
	}

	err = os.Chmod(ex, fi.Mode())
	if err != nil {
		os.Rename(tmp, ex)
		return false
	}

	// Clean up the old executable and update archive
	os.Remove(update.Name())
	os.Remove(tmp)

	return true
}
