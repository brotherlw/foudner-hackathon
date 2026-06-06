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
  <title>Agentic Paywall Flow</title>
  <style>
    :root {
      color-scheme: light;
      font-family: Arial, sans-serif;
      background: #f5f6f8;
      color: #17191c;
    }
    * { box-sizing: border-box; }
    body { margin: 0; padding: 20px; }
    main { max-width: 1360px; margin: 0 auto; }
    header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
      margin-bottom: 16px;
    }
    h1 { font-size: 24px; margin: 0; letter-spacing: 0; }
    button {
      border: 1px solid #1f2329;
      background: #1f2329;
      color: #fff;
      min-height: 40px;
      padding: 0 14px;
      border-radius: 6px;
      cursor: pointer;
      font-weight: 700;
    }
    button:disabled { opacity: 0.55; cursor: not-allowed; }
    .layout {
      display: grid;
      grid-template-columns: minmax(0, 1.2fr) minmax(360px, 0.8fr);
      gap: 16px;
      align-items: start;
    }
    section {
      background: #fff;
      border: 1px solid #d6dce5;
      border-radius: 8px;
      padding: 14px;
    }
    h2 { margin: 0 0 12px; font-size: 15px; }
    .board {
      display: grid;
      grid-template-columns: repeat(4, minmax(0, 1fr));
      gap: 12px;
      margin-bottom: 14px;
    }
    .lane {
      border: 1px solid #d8dde6;
      border-radius: 8px;
      min-height: 150px;
      padding: 12px;
      background: #fbfcfe;
      transition: border-color 160ms ease, background 160ms ease, transform 160ms ease;
    }
    .lane.active {
      border-color: #0b6bcb;
      background: #eef6ff;
      transform: translateY(-1px);
    }
    .lane.done { border-color: #18864b; background: #f0fbf5; }
    .lane.error { border-color: #b42318; background: #fff4f2; }
    .lane h3 {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 8px;
      margin: 0 0 8px;
      font-size: 14px;
    }
    .badge {
      display: inline-flex;
      align-items: center;
      min-height: 22px;
      padding: 0 8px;
      border-radius: 999px;
      background: #e8ecf2;
      color: #2d3440;
      font-size: 12px;
      white-space: nowrap;
    }
    .lane.active .badge { background: #cfe7ff; color: #064f95; }
    .lane.done .badge { background: #d8f3e3; color: #11653a; }
    .lane.error .badge { background: #ffd9d4; color: #9f1c12; }
    .lane p {
      margin: 0;
      color: #4b5563;
      line-height: 1.4;
      font-size: 13px;
    }
    .timeline {
      display: grid;
      grid-template-columns: 30px 1fr;
      gap: 8px;
      margin-top: 12px;
    }
    .step {
      display: contents;
    }
    .dot {
      width: 24px;
      height: 24px;
      border-radius: 50%;
      display: grid;
      place-items: center;
      background: #e8ecf2;
      color: #3b4552;
      font-size: 12px;
      font-weight: 700;
    }
    .step.active .dot { background: #0b6bcb; color: #fff; }
    .step.done .dot { background: #18864b; color: #fff; }
    .step.error .dot { background: #b42318; color: #fff; }
    .stepText {
      min-height: 32px;
      padding-bottom: 10px;
      font-size: 13px;
      line-height: 1.35;
      color: #3b4552;
    }
    .stepText code { font-family: Consolas, Menlo, monospace; }
    .panels {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 12px;
      margin-top: 12px;
    }
    pre {
      min-height: 300px;
      max-height: 520px;
      overflow: auto;
      white-space: pre-wrap;
      word-break: break-word;
      background: #111418;
      color: #e8edf3;
      border-radius: 6px;
      padding: 14px;
      margin: 0;
      line-height: 1.45;
      font: 13px Consolas, Menlo, monospace;
    }
    .terminal pre { min-height: 642px; }
    .hint {
      margin: 0 0 12px;
      color: #5b6572;
      font-size: 13px;
      line-height: 1.4;
    }
    @media (max-width: 980px) {
      .layout, .panels { grid-template-columns: 1fr; }
      .board { grid-template-columns: 1fr 1fr; }
      .terminal pre { min-height: 320px; }
    }
    @media (max-width: 620px) {
      body { padding: 14px; }
      header { align-items: flex-start; flex-direction: column; }
      .board { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <main>
    <header>
      <h1>Agentic Paywall Flow</h1>
      <button id="run">Run Flow</button>
    </header>
    <div class="layout">
      <section>
        <h2>Gateway to Agent to Publisher</h2>
        <div class="board">
          <div class="lane" id="agent">
            <h3>Agent <span class="badge">idle</span></h3>
            <p>Requests paid publisher content, pays, receives a grant, and retries.</p>
          </div>
          <div class="lane" id="gateway">
            <h3>Gateway <span class="badge">idle</span></h3>
            <p>Issues 402 challenges, verifies paid status, signs grants, and enforces quota.</p>
          </div>
          <div class="lane" id="payment">
            <h3>Payment <span class="badge">idle</span></h3>
            <p>Mock/Mollie test path completes EUR payment before grant issuance.</p>
          </div>
          <div class="lane" id="publisher">
            <h3>Publisher <span class="badge">idle</span></h3>
            <p>Premium content is served only after the gateway accepts a valid grant.</p>
          </div>
        </div>

        <div class="timeline" id="timeline">
          <div class="step" data-step="residency"><div class="dot">1</div><div class="stepText">Gateway publishes EU residency statement at <code>/.well-known/data-residency</code>.</div></div>
          <div class="step" data-step="challenge"><div class="dot">2</div><div class="stepText">Agent requests publisher content without a grant and receives <code>402</code>.</div></div>
          <div class="step" data-step="initiate"><div class="dot">3</div><div class="stepText">Agent initiates a EUR payment through <code>/pay/initiate</code>.</div></div>
          <div class="step" data-step="complete"><div class="dot">4</div><div class="stepText">Payment test path completes payment via <code>/pay/complete-test</code>.</div></div>
          <div class="step" data-step="verify"><div class="dot">5</div><div class="stepText">Agent polls <code>/grants/verify</code> and receives a grant.</div></div>
          <div class="step" data-step="content"><div class="dot">6</div><div class="stepText">Gateway accepts <code>PAYMENT-GRANT</code> and serves publisher content.</div></div>
          <div class="step" data-step="quota"><div class="dot">7</div><div class="stepText">Same grant is reused and rejected with <code>402</code> after quota.</div></div>
        </div>

        <div class="panels">
          <section>
            <h2>Latest Response</h2>
            <pre id="response">Open DevTools Network, then run the flow.</pre>
          </section>
          <section>
            <h2>Demo Ledger</h2>
            <pre id="ledgerView">Events will appear here as the browser flow runs.</pre>
          </section>
        </div>
      </section>

      <section class="terminal">
        <h2>Gateway Interaction Feed</h2>
        <p class="hint">This mirrors the terminal story: payment IDs, amounts, grant status, and quota outcome. Full grant tokens and IPs are not shown.</p>
        <pre id="feed"></pre>
      </section>
    </div>
  </main>
  <script>
    const run = document.querySelector("#run");
    const feed = document.querySelector("#feed");
    const response = document.querySelector("#response");
    const ledgerView = document.querySelector("#ledgerView");
    const lanes = {
      agent: document.querySelector("#agent"),
      gateway: document.querySelector("#gateway"),
      payment: document.querySelector("#payment"),
      publisher: document.querySelector("#publisher")
    };
    const steps = new Map([...document.querySelectorAll(".step")].map(el => [el.dataset.step, el]));
    const ledger = [];

    function setLane(id, state, text) {
      const lane = lanes[id];
      lane.classList.remove("active", "done", "error");
      if (state) lane.classList.add(state);
      lane.querySelector(".badge").textContent = text;
    }

    function setStep(id, state) {
      const step = steps.get(id);
      step.classList.remove("active", "done", "error");
      if (state) step.classList.add(state);
    }

    function line(text) {
      const ts = new Date().toLocaleTimeString();
      feed.textContent += "[" + ts + "] " + text + "\n";
      feed.scrollTop = feed.scrollHeight;
    }

    function show(title, value) {
      const text = typeof value === "string" ? value : JSON.stringify(value, null, 2);
      response.textContent = "==> " + title + "\n" + text;
    }

    function ledgerEvent(event) {
      ledger.push({ ts: new Date().toISOString(), ...event });
      ledgerView.textContent = ledger.map(item => JSON.stringify(item)).join("\n");
      ledgerView.scrollTop = ledgerView.scrollHeight;
    }

    async function bodyFor(resp) {
      const contentType = resp.headers.get("content-type") || "";
      if (contentType.includes("application/json")) return resp.json();
      return resp.text();
    }

    async function runStep(stepID, laneID, label, fn) {
      setStep(stepID, "active");
      setLane(laneID, "active", "working");
      line(label);
      try {
        const result = await fn();
        setStep(stepID, "done");
        setLane(laneID, "done", "ok");
        return result;
      } catch (err) {
        setStep(stepID, "error");
        setLane(laneID, "error", "error");
        line("error " + (err.message || String(err)));
        throw err;
      }
    }

    async function runFlow() {
      run.disabled = true;
      feed.textContent = "";
      response.textContent = "";
      ledger.length = 0;
      ledgerView.textContent = "";
      for (const key of Object.keys(lanes)) setLane(key, "", "idle");
      for (const step of steps.values()) step.classList.remove("active", "done", "error");

      try {
        const residency = await runStep("residency", "gateway", "GET /.well-known/data-residency", async () => {
          const resp = await fetch("/.well-known/data-residency");
          const body = await bodyFor(resp);
          show("Residency " + resp.status, body);
          line("gateway residency region=" + body.region + " raw_ip_retained=" + body.raw_ip_retained + " cross_border_transfer=" + body.cross_border_transfer);
          return body;
        });

        const challenge = await runStep("challenge", "agent", "GET /api/premium-report without grant", async () => {
          setLane("gateway", "active", "402");
          const resp = await fetch("/api/premium-report");
          const body = await bodyFor(resp);
          show("Initial content request " + resp.status, body);
          line("gateway challenge_issued resource=/api/premium-report amount=0.50 currency=EUR status=" + resp.status);
          ledgerEvent({ type: "challenge_issued", resource_path: "/api/premium-report", amount: "0.50", currency: "EUR", decision: "denied" });
          return body;
        });

        const payment = await runStep("initiate", "payment", "POST /pay/initiate amount=0.50 EUR", async () => {
          const resp = await fetch("/pay/initiate", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ resource_path: "/api/premium-report", amount: "0.50", currency: "EUR" })
          });
          const body = await bodyFor(resp);
          show("Payment initiated " + resp.status, body);
          line("gateway payment_initiated payment_id=" + body.payment_id + " resource=/api/premium-report amount=0.50 currency=EUR status=" + body.status);
          ledgerEvent({ type: "payment_initiated", payment_id: body.payment_id, resource_path: "/api/premium-report", amount: "0.50", currency: "EUR", decision: body.status });
          return body;
        });

        await runStep("complete", "payment", "POST /pay/complete-test payment_id=" + payment.payment_id, async () => {
          const resp = await fetch("/pay/complete-test", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ payment_id: payment.payment_id })
          });
          const body = await bodyFor(resp);
          show("Payment completed " + resp.status, body);
          line("gateway payment_paid payment_id=" + payment.payment_id + " amount=0.50 currency=EUR");
          line("gateway grant_issued payment_id=" + payment.payment_id + " quota=1");
          ledgerEvent({ type: "payment_paid", payment_id: payment.payment_id, resource_path: "/api/premium-report", amount: "0.50", currency: "EUR", decision: "paid" });
          ledgerEvent({ type: "grant_issued", payment_id: payment.payment_id, resource_path: "/api/premium-report", amount: "0.50", currency: "EUR", decision: "granted" });
          return body;
        });

        const grant = await runStep("verify", "agent", "GET /grants/verify payment_id=" + payment.payment_id, async () => {
          const resp = await fetch("/grants/verify?payment_id=" + encodeURIComponent(payment.payment_id));
          const body = await bodyFor(resp);
          show("Grant verified " + resp.status, { ready: body.ready, access_grant_preview: body.access_grant ? body.access_grant.slice(0, 18) + "..." : "" });
          line("gateway grant_verify payment_id=" + payment.payment_id + " ready=" + body.ready);
          return body;
        });

        await runStep("content", "publisher", "GET /api/premium-report with PAYMENT-GRANT", async () => {
          const resp = await fetch("/api/premium-report", { headers: { "PAYMENT-GRANT": grant.access_grant } });
          const body = await bodyFor(resp);
          show("Content with grant " + resp.status, body);
          line("gateway access_granted payment_id=" + payment.payment_id + " amount=0.50 currency=EUR quota=1");
          ledgerEvent({ type: "access_granted", resource_path: "/api/premium-report", amount: "0.50", currency: "EUR", decision: "granted" });
          return body;
        });

        await runStep("quota", "gateway", "GET /api/premium-report with same grant again", async () => {
          const resp = await fetch("/api/premium-report", { headers: { "PAYMENT-GRANT": grant.access_grant } });
          const body = await bodyFor(resp);
          show("Same grant after quota " + resp.status, body);
          line("gateway access_denied payment_id=" + payment.payment_id + " reason=grant quota exhausted status=" + resp.status);
          ledgerEvent({ type: "access_denied", resource_path: "/api/premium-report", amount: "0.50", currency: "EUR", decision: "denied" });
          ledgerEvent({ type: "challenge_issued", resource_path: "/api/premium-report", amount: "0.50", currency: "EUR", decision: "denied" });
          return body;
        });

        setLane("agent", "done", "complete");
        setLane("gateway", "done", "complete");
        setLane("payment", "done", "complete");
        setLane("publisher", "done", "complete");
        line("flow complete residency=" + residency.region);
      } catch (err) {
        show("Error", err.message || String(err));
      } finally {
        run.disabled = false;
      }
    }

    run.addEventListener("click", runFlow);
  </script>
</body>
</html>`
