package langs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type RustLangHelper struct {
	BaseHelper
	Version string
}

func (h *RustLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *RustLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (h *RustLangHelper) CustomMemory() uint64 {
	return 0
}

func (lh *RustLangHelper) LangStrings() []string {
	return []string{"rust", fmt.Sprintf("rust%s", lh.Version)}
}
func (lh *RustLangHelper) Extensions() []string {
	return []string{".rs"}
}

func (lh *RustLangHelper) BuildFromImage() (string, error) {
	return fmt.Sprintf("fnproject/rust:%s-dev", lh.Version), nil
}

func (lh *RustLangHelper) RunFromImage() (string, error) {
	return fmt.Sprintf("fnproject/rust:%s", lh.Version), nil
}

func (h *RustLangHelper) DockerfileBuildCmds() []string {
	return []string{
		"COPY . .",
		"RUN cargo build --release",
	}
}

func (h *RustLangHelper) DockerfileCopyCmds() []string {
	return []string{
		"COPY --from=build-stage /function/target/release/func /function/",
	}
}

func (lh *RustLangHelper) Entrypoint() (string, error) {
	return "./func", nil
}

func (lh *RustLangHelper) HasBoilerplate() bool { return true }

func (lh *RustLangHelper) GenerateBoilerplate(path string) error {
	srcDir := filepath.Join(path, "src")
	if err := os.Mkdir(srcDir, os.FileMode(0755)); err != nil {
		return err
	}

	codeFile := filepath.Join(srcDir, "main.rs")
	if exists(codeFile) {
		return errors.New("A rust project already exists, cancelling init")
	}
	if err := ioutil.WriteFile(codeFile, []byte(helloRustSrcBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}
	cargoToml := "Cargo.toml"
	fdkVersion, _ := lh.GetLatestFDKVersion()
	if err := ioutil.WriteFile(cargoToml, []byte(fmt.Sprintf(cargoTomlBoilerplate, fdkVersion)), os.FileMode(0644)); err != nil {
		return err
	}

	return nil
}

func (h *RustLangHelper) GetLatestFDKVersion() (string, error) {
	// Github API has limit on number of calls
	resp, err := http.Get("https://api.github.com/repos/fnproject/fdk-rust/tags")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	responseBody := []githubTagResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", err
	}
	if len(responseBody) == 0 {
		return "", errors.New("Could not read latest version of FDK from tags")
	}
	return responseBody[0].Name, nil
}

const (
	helloRustSrcBoilerplate = `use fdk::{Function, FunctionError, RuntimeContext};
use tokio; // Tokio for handling future.

#[tokio::main]
async fn main() -> Result<(), FunctionError> {
    if let Err(e) = Function::run(|_: &mut RuntimeContext, i: String| {
        Ok(format!(
            "Hello {}!",
            if i.is_empty() {
                "world"
            } else {
                i.trim_end_matches("\n")
            }
        ))
    })
    .await
    {
        eprintln!("{}", e);
    }
    Ok(())
}
`

	cargoTomlBoilerplate = `[package]
name = "func"
version = "0.1.0"
edition = "2018"

# See more keys and their definitions at https://doc.rust-lang.org/cargo/reference/manifest.html

[dependencies]
tokio = { version = "1.6", features = ["macros", "rt-multi-thread"] }
serde = { version = "1", features = ["derive"] }
fdk = "%s"
`
)

func (h *RustLangHelper) FixImagesOnInit() bool {
	return true
}
