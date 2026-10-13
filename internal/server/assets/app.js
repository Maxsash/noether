"use strict";

const reposEl = document.getElementById("repos");
const detailEl = document.getElementById("detail");
const wide = window.matchMedia("(min-width: 761px)");

let summaries = [];
let loading = 0;

// h builds an element. The html prop is only ever given markdown the server rendered with
// raw HTML stripped; every other value is set as text.
function h(tag, props = {}, ...children) {
  const el = document.createElement(tag);
  for (const [key, value] of Object.entries(props)) {
    if (value === undefined || value === null || value === false) continue;
    if (key === "class") el.className = value;
    else if (key === "html") el.innerHTML = value;
    else el.setAttribute(key, value === true ? "" : value);
  }
  for (const child of children.flat()) {
    if (child !== null && child !== undefined && child !== false) el.append(child);
  }
  return el;
}

const relative = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });
const units = [
  ["year", 31536000],
  ["month", 2592000],
  ["week", 604800],
  ["day", 86400],
  ["hour", 3600],
  ["minute", 60],
];

function ago(iso) {
  if (!iso) return "never";
  const seconds = (new Date(iso) - Date.now()) / 1000;
  for (const [unit, size] of units) {
    if (Math.abs(seconds) >= size) return relative.format(Math.round(seconds / size), unit);
  }
  return "just now";
}

function dayLabel(iso) {
  return new Date(iso).toLocaleDateString(undefined, { weekday: "short", day: "numeric", month: "short", year: "numeric" });
}

function plural(n, noun) {
  return `${n} ${noun}${n === 1 ? "" : "s"}`;
}

function currentId() {
  return decodeURIComponent(location.hash.replace(/^#\/?/, ""));
}

// On a wide screen the most recently active repository opens when none is chosen.
function selectedId() {
  return currentId() || (wide.matches && summaries.length ? summaries[0].id : "");
}

function repoHref(id) {
  return `#/${encodeURIComponent(id)}`;
}

async function getJSON(url) {
  const response = await fetch(url);
  if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
  return response.json();
}

function renderList(list) {
  summaries = list.repos;
  const problems = list.problems?.length ? h("p", { class: "problems" }, list.problems.join(" · ")) : null;
  reposEl.replaceChildren(
    h(
      "ul",
      {},
      summaries.map((s) =>
        h(
          "li",
          {},
          h(
            "a",
            { class: "repo", href: repoHref(s.id) },
            h("span", { class: "repo-name" }, s.name),
            s.id !== s.name ? h("span", { class: "repo-where" }, s.id) : null,
            h("span", { class: "repo-meta" }, listMeta(s)),
          ),
        ),
      ),
    ),
  );
  if (problems) reposEl.append(problems);
}

function listMeta(s) {
  const parts = [];
  if (s.warnings) parts.push(`⚠ ${s.warnings}`);
  if (s.due) parts.push(`${s.due} due`);
  if (s.openTasks) parts.push(`${s.openTasks} open`);
  parts.push(ago(s.lastActivity));
  return parts.join(" · ");
}

function section(title, ...content) {
  return h("section", { "aria-label": title }, h("h3", {}, title), ...content);
}

function openLink(href, label = "open") {
  return href ? h("a", { class: "open-link", href }, label) : null;
}

function header(d) {
  const facts = [d.branch ? `on ${d.branch}` : "detached HEAD"];
  if (d.changes?.length) facts.push(`${plural(d.changes.length, "uncommitted file")}`);
  if (d.lastActivity) facts.push(`active ${ago(d.lastActivity)}`);
  return h(
    "div",
    { class: "title" },
    h("a", { class: "back", href: "#/" }, "← All repositories"),
    h("h2", {}, d.name),
    h("p", { class: "mono" }, d.path),
    h("p", {}, facts.join(" · ")),
  );
}

function attentionView(notices) {
  if (!notices?.length) return h("p", { class: "muted" }, "Nothing needs attention.");
  return h(
    "ul",
    { class: "notices" },
    notices.map((n) => h("li", { class: n.level }, n.text)),
  );
}

function stateView(d) {
  if (!d.state) {
    return h("p", { class: "muted" }, "No state file. Name one in this repository's noether block, or add docs/handoff.md.");
  }
  const prose = h("div", { class: "prose", html: d.stateHtml });
  const toggle = h("button", { type: "button" }, "Show all");
  toggle.addEventListener("click", () => {
    toggle.textContent = prose.classList.toggle("open") ? "Show less" : "Show all";
  });
  return h(
    "div",
    { class: "card" },
    h(
      "div",
      { class: "doc-meta" },
      h("span", { class: "mono" }, d.state.path),
      h("span", {}, plural(d.state.lines, "line")),
      h("span", {}, `updated ${ago(d.state.updated)}`),
      openLink(d.stateOpen),
    ),
    prose,
    d.state.lines > 40 ? toggle : null,
  );
}

function groupTasks(tasks) {
  const groups = new Map();
  for (const t of tasks) {
    const key = `${t.file}\u0000${t.section}`;
    if (!groups.has(key)) groups.set(key, { file: t.file, section: t.section, tasks: [] });
    groups.get(key).tasks.push(t);
  }
  return [...groups.values()];
}

function taskList(tasks) {
  return h(
    "ul",
    { class: "tasks" },
    tasks.map((t) =>
      h("li", { class: t.done ? "done" : null }, h("span", { class: "task-text", html: t.html }), openLink(t.open)),
    ),
  );
}

function tasksView(tasks) {
  if (!tasks.length) {
    return h("p", { class: "muted" }, "No task lists found. Checkboxes (- [ ]) in the files named as tasks show here.");
  }
  const open = tasks.filter((t) => !t.done);
  const done = tasks.filter((t) => t.done);
  return h(
    "div",
    { class: "card" },
    open.length
      ? groupTasks(open).map((g) =>
          h("div", { class: "group" }, h("h4", {}, [g.section, g.file].filter(Boolean).join(" — ")), taskList(g.tasks)),
        )
      : h("p", { class: "muted" }, "Everything is done."),
    done.length ? h("details", {}, h("summary", {}, `${done.length} done`), taskList(done)) : null,
  );
}

function cadenceLabel(c) {
  return c.commits ? `every ${c.every} or ${c.commits} commits` : `every ${c.every}`;
}

function upkeepView(jobs) {
  if (!jobs?.length) {
    return h(
      "p",
      { class: "muted" },
      "No upkeep jobs are scheduled. Add an [upkeep] table to this repository's noether block or to the defaults in your config.",
    );
  }
  return h(
    "div",
    { class: "card" },
    h(
      "table",
      { class: "grid" },
      h("thead", {}, h("tr", {}, h("th", {}, "Job"), h("th", {}, "Schedule"), h("th", {}, "Last run"), h("th", {}, "Status"))),
      h(
        "tbody",
        {},
        jobs.map((j) =>
          h(
            "tr",
            {},
            h("td", {}, h("strong", {}, j.kind)),
            h("td", {}, cadenceLabel(j.cadence)),
            h(
              "td",
              {},
              j.last ? `${ago(j.last)}${j.summary ? ` — ${j.summary}` : ""}` : "never",
              j.commitsSince ? h("div", { class: "muted" }, `${plural(j.commitsSince, "commit")} since`) : null,
            ),
            h("td", {}, h("span", { class: `badge ${j.state}`, title: j.reason || null }, j.state)),
          ),
        ),
      ),
    ),
  );
}

function historyView(events) {
  if (!events?.length) return h("p", { class: "muted" }, "No history yet.");
  const items = [];
  let lastDay = "";
  for (const e of events) {
    const day = dayLabel(e.when);
    if (day !== lastDay) {
      items.push(h("li", { class: "day" }, day));
      lastDay = day;
    }
    items.push(
      h(
        "li",
        { class: `event ${e.kind}` },
        h("span", { class: e.kind === "commit" ? "ref mono" : "ref" }, e.ref),
        h("span", {}, e.title),
      ),
    );
  }
  return h("div", { class: "card" }, h("ul", { class: "history" }, items));
}

function docsView(docs) {
  if (!docs.length) return h("p", { class: "muted" }, "No markdown files at the top level or in docs/.");
  return h(
    "div",
    { class: "card" },
    h(
      "table",
      { class: "grid" },
      h("thead", {}, h("tr", {}, h("th", {}, "File"), h("th", {}, "Lines"), h("th", {}, "Updated"))),
      h(
        "tbody",
        {},
        docs.map((d) =>
          h(
            "tr",
            {},
            h("td", {}, d.open ? h("a", { href: d.open, class: "mono" }, d.path) : h("span", { class: "mono" }, d.path)),
            h("td", {}, String(d.lines)),
            h("td", {}, ago(d.updated)),
          ),
        ),
      ),
    ),
  );
}

function renderDetail(d) {
  detailEl.replaceChildren(
    header(d),
    section("Needs attention", attentionView(d.attention)),
    section("Now", stateView(d)),
    section("To do", tasksView(d.tasks)),
    section("Upkeep", upkeepView(d.upkeep)),
    section("History", historyView(d.history)),
    section("Docs", docsView(d.docs)),
  );
}

function showMessage(el, text) {
  el.replaceChildren(h("p", { class: "muted" }, text));
}

function markCurrent(id) {
  for (const link of reposEl.querySelectorAll(".repo")) {
    if (link.getAttribute("href") === repoHref(id)) link.setAttribute("aria-current", "page");
    else link.removeAttribute("aria-current");
  }
  document.body.classList.toggle("has-detail", Boolean(currentId()));
}

async function loadDetail({ focus = false } = {}) {
  const ticket = ++loading;
  const id = selectedId();
  markCurrent(id);
  if (!id) {
    showMessage(detailEl, summaries.length ? "Choose a repository." : "No repositories found under the configured roots.");
    return;
  }
  try {
    const detail = await getJSON(`/api/repos/${encodeURIComponent(id)}`);
    if (ticket !== loading) return;
    renderDetail(detail);
    if (focus) detailEl.focus();
  } catch (err) {
    if (ticket === loading) showMessage(detailEl, `Could not read ${id}: ${err.message}`);
  }
}

async function loadAll() {
  try {
    renderList(await getJSON("/api/repos"));
  } catch (err) {
    showMessage(reposEl, `Could not read repositories: ${err.message}`);
    return;
  }
  await loadDetail();
}

window.addEventListener("hashchange", () => loadDetail({ focus: true }));
document.getElementById("refresh").addEventListener("click", loadAll);
document.addEventListener("visibilitychange", () => {
  if (!document.hidden) loadAll();
});
loadAll();
