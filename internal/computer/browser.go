package computer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type BrowserManager struct {
	active   bool
	ctx      context.Context
	cancel   context.CancelFunc
	endpoint string
}

func NewBrowserManager() *BrowserManager {
	return &BrowserManager{}
}

func (bm *BrowserManager) Start(parent context.Context) error {
	bm.ctx, bm.cancel = context.WithCancel(parent)

	if err := bm.checkPlaywright(); err != nil {
		return fmt.Errorf("playwright not available: %w", err)
	}

	bm.active = true
	return nil
}

func (bm *BrowserManager) Stop() error {
	if bm.cancel != nil {
		bm.cancel()
	}
	bm.active = false
	return nil
}

func (bm *BrowserManager) checkPlaywright() error {
	cmd := exec.Command("npx", "playwright", "--version")
	if err := cmd.Run(); err != nil {
		cmd2 := exec.Command("npx.cmd", "playwright", "--version")
		return cmd2.Run()
	}
	return nil
}

func (bm *BrowserManager) Navigate(ctx context.Context, url string) (string, error) {
	script := bm.scriptNavigate(url)
	out, err := bm.runScript(script)
	if err != nil {
		return "", fmt.Errorf("playwright navigate: %w", err)
	}
	return out, nil
}

func (bm *BrowserManager) Click(ctx context.Context, selector string) error {
	script := bm.scriptClick(selector)
	_, err := bm.runScript(script)
	return err
}

func (bm *BrowserManager) Type(ctx context.Context, selector, text string) error {
	script := bm.scriptType(selector, text)
	_, err := bm.runScript(script)
	return err
}

func (bm *BrowserManager) Scroll(ctx context.Context, selector string, deltaY int) error {
	script := bm.scriptScroll(selector, deltaY)
	_, err := bm.runScript(script)
	return err
}

func (bm *BrowserManager) Screenshot(ctx context.Context) (string, error) {
	script := bm.scriptScreenshot()
	return bm.runScript(script)
}

func (bm *BrowserManager) PreClickSnapshot(ctx context.Context, selector string) (screenshot, elementHTML string, err error) {
	script := bm.scriptPreClick(selector)
	result, err := bm.runScript(script)
	if err != nil {
		return "", "", fmt.Errorf("pre-click snapshot: %w", err)
	}

	var data struct {
		Screenshot string `json:"screenshot"`
		Element    string `json:"element"`
	}
	if json.Unmarshal([]byte(result), &data) == nil {
		return data.Screenshot, data.Element, nil
	}
	return result, "", nil
}

func (bm *BrowserManager) Evaluate(ctx context.Context, expression string) (string, error) {
	script := bm.scriptEvaluate(expression)
	return bm.runScript(script)
}

func (bm *BrowserManager) PageInfo(ctx context.Context) (title, url string, err error) {
	script := bm.scriptPageInfo()
	result, err := bm.runScript(script)
	if err != nil {
		return "", "", err
	}
	var info struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if json.Unmarshal([]byte(result), &info) == nil {
		return info.Title, info.URL, nil
	}
	return "", "", fmt.Errorf("parse page info: %s", result)
}

func (bm *BrowserManager) runScript(script string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "curse-browser-*")
	if err != nil {
		return "", fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	scriptPath := filepath.Join(tmpDir, "script.mjs")
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return "", fmt.Errorf("write script: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "npx.cmd", "playwright", "run", scriptPath)
	default:
		cmd = exec.CommandContext(ctx, "npx", "playwright", "run", scriptPath)
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stderr.String(), fmt.Errorf("playwright exec: %w\n%s", err, stderr.String())
	}

	return strings.TrimSpace(out.String()), nil
}

func (bm *BrowserManager) scriptNavigate(url string) string {
	return fmt.Sprintf(`import { chromium } from 'playwright';
(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  await page.goto('%s', { waitUntil: 'networkidle' });
  const screenshot = await page.screenshot({ type: 'png', fullPage: true });
  console.log(screenshot.toString('base64'));
  await browser.close();
})();`, url)
}

func (bm *BrowserManager) scriptClick(selector string) string {
	escaped := strings.ReplaceAll(selector, "'", "\\'")
	return fmt.Sprintf(`import { chromium } from 'playwright';
(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  const el = await page.waitForSelector('%s');
  await el.click();
  await browser.close();
})();`, escaped)
}

func (bm *BrowserManager) scriptType(selector, text string) string {
	escapedSel := strings.ReplaceAll(selector, "'", "\\'")
	escapedText := strings.ReplaceAll(text, "'", "\\'")
	return fmt.Sprintf(`import { chromium } from 'playwright';
(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  const el = await page.waitForSelector('%s');
  await el.fill('%s');
  await browser.close();
})();`, escapedSel, escapedText)
}

func (bm *BrowserManager) scriptScroll(selector string, deltaY int) string {
	escaped := strings.ReplaceAll(selector, "'", "\\'")
	return fmt.Sprintf(`import { chromium } from 'playwright';
(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  if ('%s' === 'window') {
    await page.evaluate(() => window.scrollBy(0, %d));
  } else {
    const el = await page.waitForSelector('%s');
    await el.evaluate((e) => e.scrollBy(0, %d));
  }
  await browser.close();
})();`, escaped, deltaY, escaped, deltaY)
}

func (bm *BrowserManager) scriptScreenshot() string {
	return `import { chromium } from 'playwright';
(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  const screenshot = await page.screenshot({ type: 'png', fullPage: true });
  console.log(screenshot.toString('base64'));
  await browser.close();
})();`
}

func (bm *BrowserManager) scriptPreClick(selector string) string {
	escaped := strings.ReplaceAll(selector, "'", "\\'")
	return fmt.Sprintf(`import { chromium } from 'playwright';
(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  const screenshot = await page.screenshot({ type: 'png' });
  let element = '';
  try {
    const el = await page.waitForSelector('%s', { timeout: 5000 });
    element = await el.evaluate(e => e.outerHTML.substring(0, 500));
  } catch(e) { element = ''; }
  console.log(JSON.stringify({ screenshot: screenshot.toString('base64'), element: element }));
  await browser.close();
})();`, escaped)
}

func (bm *BrowserManager) scriptEvaluate(expression string) string {
	escaped := strings.ReplaceAll(expression, "'", "\\'")
	return fmt.Sprintf(`import { chromium } from 'playwright';
(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  const result = await page.evaluate(() => { return %s; });
  console.log(JSON.stringify(result));
  await browser.close();
})();`, escaped)
}

func (bm *BrowserManager) scriptPageInfo() string {
	return `import { chromium } from 'playwright';
(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  const info = { title: await page.title(), url: page.url() };
  console.log(JSON.stringify(info));
  await browser.close();
})();`
}
