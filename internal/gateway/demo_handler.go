package gateway

import "net/http"

func BrowserDemoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(browserDemoHTML))
	})
}

const browserDemoHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Examplepedia - Protected Article</title>
  <style>
    :root {
      color-scheme: light;
      font-family: Georgia, "Times New Roman", serif;
      background: #f8f9fa;
      color: #202122;
    }
    * { box-sizing: border-box; }
    body { margin: 0; }
    .topbar {
      height: 52px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
      padding: 0 22px;
      border-bottom: 1px solid #c8ccd1;
      background: #fff;
    }
    .brand {
      font-family: Arial, sans-serif;
      font-weight: 800;
      font-size: 22px;
      letter-spacing: 0;
    }
    .search {
      width: min(360px, 42vw);
      height: 32px;
      border: 1px solid #a2a9b1;
      padding: 0 10px;
      font-family: Arial, sans-serif;
    }
    main {
      max-width: 980px;
      margin: 0 auto;
      padding: 28px 24px 60px;
      background: #fff;
      min-height: calc(100vh - 52px);
      border-left: 1px solid #eaecf0;
      border-right: 1px solid #eaecf0;
    }
    h1 {
      margin: 0 0 6px;
      font-size: 34px;
      font-weight: 400;
      letter-spacing: 0;
      border-bottom: 1px solid #a2a9b1;
      padding-bottom: 8px;
    }
    .subtitle {
      font-family: Arial, sans-serif;
      color: #54595d;
      margin-bottom: 18px;
    }
    .notice {
      border: 1px solid #a2a9b1;
      border-left: 4px solid #36c;
      background: #f8f9fa;
      padding: 12px;
      margin: 16px 0;
      font-family: Arial, sans-serif;
      line-height: 1.45;
    }
    .protected {
      border: 1px solid #c8ccd1;
      background: #f8f9fa;
      padding: 14px;
      margin: 18px 0;
      font-family: Arial, sans-serif;
    }
    .lock {
      display: inline-flex;
      align-items: center;
      min-height: 24px;
      padding: 0 8px;
      border-radius: 999px;
      background: #fee7e6;
      color: #b32424;
      font-size: 12px;
      font-weight: 700;
      margin-left: 8px;
      vertical-align: middle;
    }
    p {
      line-height: 1.62;
      font-size: 16px;
      margin: 0 0 14px;
    }
    .redacted {
      display: block;
      height: 13px;
      border-radius: 3px;
      background: linear-gradient(90deg, #c8ccd1, #eaecf0);
      margin: 10px 0;
    }
    code, pre {
      font-family: Consolas, Menlo, monospace;
    }
    pre {
      white-space: pre-wrap;
      border: 1px solid #c8ccd1;
      background: #f8f9fa;
      padding: 12px;
      overflow: auto;
      font-size: 13px;
    }
    .browser-feed {
      min-height: 180px;
      max-height: 280px;
      background: #101418;
      color: #f4f7fb;
      border-color: #101418;
    }
    .small {
      font-family: Arial, sans-serif;
      font-size: 13px;
      color: #54595d;
    }
  </style>
</head>
<body>
  <div class="topbar">
    <div class="brand">Examplepedia</div>
    <input class="search" value="European market infrastructure" aria-label="Search">
  </div>
  <main>
    <h1>European Market Infrastructure Report <span class="lock">Protected</span></h1>
    <div class="subtitle">From Examplepedia, the customer publisher demo site</div>

    <div class="notice">
      This page represents the customer website. The protected article body is not served to agents
      until the gateway verifies payment and accepts an access grant.
    </div>

    <p>
      European market infrastructure refers to systems and institutions that coordinate clearing,
      settlement, payment routing, and access to regulated market information across European venues.
    </p>
    <p>
      Human readers see this article shell. Machine clients requesting the premium report endpoint are
      routed through the local gateway and receive a machine-readable HTTP 402 challenge.
    </p>

    <div class="protected">
      <strong>Premium report body</strong>
      <span class="redacted"></span>
      <span class="redacted" style="width: 91%"></span>
      <span class="redacted" style="width: 76%"></span>
      <span class="redacted" style="width: 84%"></span>
      <p class="small">Protected resource: <code>/api/premium-report</code></p>
    </div>

    <h2>Developer Console Probe</h2>
    <p class="small">
      Open DevTools Console to watch this customer website make browser-side probe requests. You can
      also run <code>examplepediaProbe()</code> again manually.
    </p>
    <pre>examplepediaProbe()</pre>

    <h2>Browser Traffic Mirror</h2>
    <p class="small">
      This mirrors the same customer-side traffic that should appear in DevTools Console.
    </p>
    <pre id="browserFeed" class="browser-feed">waiting for browser traffic...</pre>
  </main>

  <script>
    const browserFeed = document.querySelector("#browserFeed");

    function customerLog(message, payload) {
      const line = "[" + new Date().toLocaleTimeString() + "] " + message;
      if (browserFeed.textContent === "waiting for browser traffic...") {
        browserFeed.textContent = "";
      }
      browserFeed.textContent += line + (payload === undefined ? "" : " " + JSON.stringify(payload)) + "\n";
      browserFeed.scrollTop = browserFeed.scrollHeight;
      if (payload === undefined) {
        console.warn(line);
      } else {
        console.warn(line, payload);
      }
    }

    customerLog("[customer] page loaded");
    customerLog("[customer] protected resource: /api/premium-report");

    async function logFetch(label, url, options) {
      customerLog("[customer] -> " + label + " " + url);
      const resp = await fetch(url, options);
      const contentType = resp.headers.get("content-type") || "";
      const body = contentType.includes("application/json") ? await resp.json() : await resp.text();
      customerLog("[customer] <- " + label + " status=" + resp.status, body);
      return { status: resp.status, body };
    }

    window.examplepediaProbe = async function examplepediaProbe() {
      customerLog("[customer] probe started");
      await logFetch("residency", "/.well-known/data-residency");
      await logFetch("protected_without_grant", "/api/premium-report");
      customerLog("[customer] probe complete");
    };

    window.addEventListener("load", function () {
      setTimeout(function () {
        window.examplepediaProbe().catch(function (err) {
          console.error("[customer] probe_failed", err);
        });
      }, 500);
    });
  </script>
</body>
</html>`
