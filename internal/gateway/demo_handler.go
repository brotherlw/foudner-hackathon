package gateway

import "net/http"

func BrowserDemoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(browserDemoHTML))
	})
}

const browserDemoHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Agentic Paywall Demo</title>
  <style>
    :root {
      color-scheme: light;
      font-family: Arial, sans-serif;
      background: #f6f7f9;
      color: #17191c;
    }
    body {
      margin: 0;
      padding: 24px;
    }
    main {
      max-width: 1120px;
      margin: 0 auto;
    }
    header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
      margin-bottom: 18px;
    }
    h1 {
      font-size: 24px;
      margin: 0;
    }
    button {
      border: 1px solid #20242a;
      background: #20242a;
      color: #fff;
      min-height: 40px;
      padding: 0 14px;
      border-radius: 6px;
      cursor: pointer;
      font-weight: 700;
    }
    button:disabled {
      opacity: 0.55;
      cursor: not-allowed;
    }
    .grid {
      display: grid;
      grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
      gap: 16px;
    }
    section {
      background: #fff;
      border: 1px solid #d8dde6;
      border-radius: 8px;
      padding: 16px;
    }
    h2 {
      margin: 0 0 12px;
      font-size: 15px;
    }
    ol {
      margin: 0;
      padding-left: 22px;
    }
    li {
      margin: 0 0 10px;
      line-height: 1.35;
    }
    code, pre {
      font-family: Consolas, Menlo, monospace;
    }
    pre {
      min-height: 420px;
      overflow: auto;
      white-space: pre-wrap;
      word-break: break-word;
      background: #111418;
      color: #e8edf3;
      border-radius: 6px;
      padding: 14px;
      margin: 0;
      line-height: 1.45;
      font-size: 13px;
    }
    .ok { color: #0d6b3e; }
    .warn { color: #936400; }
    .err { color: #a51d2d; }
    @media (max-width: 760px) {
      body { padding: 14px; }
      header { align-items: flex-start; flex-direction: column; }
      .grid { grid-template-columns: 1fr; }
      pre { min-height: 280px; }
    }
  </style>
</head>
<body>
  <main>
    <header>
      <h1>Agentic Paywall Browser Demo</h1>
      <button id="run">Run Flow</button>
    </header>
    <div class="grid">
      <section>
        <h2>Flow</h2>
        <ol id="steps">
          <li>GET <code>/.well-known/data-residency</code></li>
          <li>GET <code>/api/premium-report</code> without grant</li>
          <li>POST <code>/pay/initiate</code></li>
          <li>POST <code>/pay/complete-test</code></li>
          <li>GET <code>/grants/verify</code></li>
          <li>GET <code>/api/premium-report</code> with grant</li>
          <li>GET <code>/api/premium-report</code> with same grant again</li>
        </ol>
      </section>
      <section>
        <h2>Output</h2>
        <pre id="out">Open browser developer tools and keep the Network tab visible, then run the flow.</pre>
      </section>
    </div>
  </main>
  <script>
    const out = document.querySelector("#out");
    const run = document.querySelector("#run");

    function log(title, value) {
      const text = typeof value === "string" ? value : JSON.stringify(value, null, 2);
      out.textContent += "\n\n==> " + title + "\n" + text;
      out.scrollTop = out.scrollHeight;
    }

    async function bodyFor(response) {
      const contentType = response.headers.get("content-type") || "";
      if (contentType.includes("application/json")) return response.json();
      return response.text();
    }

    async function runFlow() {
      run.disabled = true;
      out.textContent = "";
      try {
        const residencyResp = await fetch("/.well-known/data-residency");
        log("Residency " + residencyResp.status, await bodyFor(residencyResp));

        const challengeResp = await fetch("/api/premium-report");
        const challenge = await bodyFor(challengeResp);
        log("Initial content request " + challengeResp.status, challenge);

        const paymentResp = await fetch("/pay/initiate", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            resource_path: "/api/premium-report",
            amount: "0.50",
            currency: "EUR"
          })
        });
        const payment = await bodyFor(paymentResp);
        log("Payment initiated " + paymentResp.status, payment);

        const completeResp = await fetch("/pay/complete-test", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ payment_id: payment.payment_id })
        });
        log("Payment completed " + completeResp.status, await bodyFor(completeResp));

        const verifyResp = await fetch("/grants/verify?payment_id=" + encodeURIComponent(payment.payment_id));
        const grant = await bodyFor(verifyResp);
        log("Grant verified " + verifyResp.status, {
          ready: grant.ready,
          access_grant_preview: grant.access_grant ? grant.access_grant.slice(0, 18) + "..." : ""
        });

        const contentResp = await fetch("/api/premium-report", {
          headers: { "PAYMENT-GRANT": grant.access_grant }
        });
        log("Content with grant " + contentResp.status, await bodyFor(contentResp));

        const quotaResp = await fetch("/api/premium-report", {
          headers: { "PAYMENT-GRANT": grant.access_grant }
        });
        log("Same grant after quota " + quotaResp.status, await bodyFor(quotaResp));
      } catch (err) {
        log("Error", err.message || String(err));
      } finally {
        run.disabled = false;
      }
    }

    run.addEventListener("click", runFlow);
  </script>
</body>
</html>`
