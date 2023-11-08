package moduleinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/rrrix/find-latest-gopkg/pkg/instance"
)

type Origin struct {
	VCS  string `json:"VCS"`
	URL  string `json:"URL"`
	Ref  string `json:"Ref"`
	Hash string `json:"Hash"`
}

type ModuleInfo struct {
	Name    string
	Version string `json:"Version"`
	Time    string `json:"Time"`
	Origin  Origin `json:"Origin"`
}

type infoParserFunc = func(module, url string, resp *http.Response) (*ModuleInfo, error)

func parseDefault(module, url string, r *http.Response) (*ModuleInfo, error) {
	defer func() {
		_ = r.Body.Close()
	}()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Error reading response body for %s: %v", url, err)
		return nil, err
	}
	err = fmt.Errorf("unexpected response for module %s (%s): %s", module, url, body)
	return nil, err
}

func parseErrJson(module, url string, resp *http.Response) (*ModuleInfo, error) {
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading response body for %s: %v", module, err)
		return nil, err
	}

	s := jsonOrBust(body, true)
	log.Errorf("%s", s)
	err = fmt.Errorf("unexpected response for %s, received HTTP %d", url, resp.StatusCode)

	return nil, err
}

func parseInfo(module, url string, resp *http.Response) (*ModuleInfo, error) {
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read and unmarshal response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading response body for %s: %v", url, err)
		return nil, err
	}

	info := &ModuleInfo{Name: module}
	err = json.Unmarshal(body, info)
	if err != nil {
		log.Errorf("Error parsing JSON for %s: %v", url, err)
		log.Debugf("%s\n%s", url, string(body))
		return nil, err
	}

	log.Debugf("%+v", info)
	return info, nil
}

func jsonOrBust(o interface{}, indent bool) string {
	indented := func() ([]byte, error) {
		return json.MarshalIndent(o, "", "  ")
	}
	compact := func() ([]byte, error) {
		return json.Marshal(o)
	}
	var b []byte
	var err error
	if indent {
		b, err = indented()
	} else {
		b, err = compact()
	}
	if err != nil {
		err = fmt.Errorf("error marshaling JSON for %v: %v", o, err)
		return err.Error()
	}
	return string(b)
}

func FindLatest(proxy, module string) (*ModuleInfo, error) {
	// Make HTTP request to proxy.golang.org or given module proxy
	url := fmt.Sprintf("%s/%s/@v/@latest", proxy, module)
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		log.Errorf("Error creating request for %s: %v", module, err)
		return nil, err
	}
	log.Debugf("HTTP GET %v", url)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Error fetching module %s: %v", module, err)
		return nil, err
	}

	var nextFunc infoParserFunc

	isJson := resp.Header.Get("Content-Type") == "application/json"
	is2xx := resp.StatusCode >= 200 && resp.StatusCode <= 299

	switch {
	case is2xx && isJson:
		nextFunc = parseInfo
	case !is2xx && isJson:
		nextFunc = parseErrJson
	default:
		nextFunc = parseDefault
	}

	return nextFunc(module, url, resp)
}

func PrintInfo(m *instance.MainContext, info ModuleInfo) {

	o := m.Options
	// Print fields based on flags
	if o.Name {
		fmt.Printf("// %s\n", info.Name)
	}
	if o.Version {
		fmt.Println(info.Version)
	}
	if o.Time {
		fmt.Println(info.Time)
	}
	if o.Repo {
		fmt.Println(info.Origin.URL)
	}
	if o.Ref {
		fmt.Println(info.Origin.Ref)
	}
	if o.Hash {
		fmt.Println(info.Origin.Hash)
	}
	if o.Dump {
		// Marshal the map to a pretty-printed JSON string
		prettyJSON, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			log.Errorf("Error marshaling JSON for %s: %v", info.Name, err)
			return
		}

		// Print the pretty-printed JSON string
		fmt.Println(string(prettyJSON))
	}
}
