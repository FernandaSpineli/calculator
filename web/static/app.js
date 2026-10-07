"use strict";

const API = "/api/v1";

const ERROR_MESSAGES = {
  division_by_zero: "Divisão por zero",
  domain_error: "Valor fora do domínio da operação",
  overflow: "Resultado grande demais",
  wrong_operand_count: "Número de operandos inválido",
  unknown_operation: "Operação desconhecida",
};

const BINARY_SYMBOLS = { add: "+", subtract: "−", multiply: "×", divide: "÷", mod: "mod", power: "^" };
const ANGLE_IN = new Set(["sin", "cos", "tan"]);
const ANGLE_OUT = new Set(["asin", "acos", "atan"]);

const $ = (id) => document.getElementById(id);
const els = {
  display: document.querySelector(".display"),
  expression: $("expression"),
  value: $("value"),
  error: $("error"),
  list: $("history-list"),
  empty: $("history-empty"),
  template: $("history-item"),
};

const state = {
  entry: "0",        // raw text of the current operand
  fresh: true,       // the next digit starts a new operand
  hasOperand: true,  // entry holds a value a binary operation can use
  pending: null,     // { op, x } waiting for its second operand
  busy: false,
  history: [],
};

const unit = () => document.querySelector('input[name="unit"]:checked').value;

// ---------- formatting ----------

function fmt(n) {
  if (!Number.isFinite(n)) return String(n);
  return Number(n.toPrecision(12)).toString().replace(".", ",");
}

function operand(n) {
  const s = fmt(n);
  return n < 0 ? `(${s})` : s;
}

function describe(op, operands, angleUnit) {
  const [a, b] = operands.map(operand);
  const deg = angleUnit === "deg";
  switch (op) {
    case "power": return `${a} ^ ${b}`;
    case "root": return `${b}√${a}`;
    case "log": return `log${b}(${a})`;
    case "sqrt": return `√${a}`;
    case "cbrt": return `∛${a}`;
    case "abs": return `|${fmt(operands[0])}|`;
    case "reciprocal": return `1/${a}`;
    case "factorial": return `${a}!`;
    case "exp": return `e^${a}`;
    case "log10": return `log₁₀(${fmt(operands[0])})`;
    case "log2": return `log₂(${fmt(operands[0])})`;
  }
  if (op in BINARY_SYMBOLS) return `${a} ${BINARY_SYMBOLS[op]} ${b}`;
  if (ANGLE_IN.has(op)) return `${op}(${fmt(operands[0])}${deg ? "°" : " rad"})`;
  if (ANGLE_OUT.has(op)) return `${op.slice(1)}⁻¹(${fmt(operands[0])})`;
  return `${op}(${fmt(operands[0])})`;
}

function describeResult(calc) {
  const r = fmt(calc.result);
  return ANGLE_OUT.has(calc.operation) && calc.angle_unit === "deg" ? `${r}°` : r;
}

// ---------- rendering ----------

function render() {
  const typing = !state.fresh;
  els.value.textContent = typing ? state.entry.replace(".", ",") : fmt(Number(state.entry));
  document.querySelectorAll("[data-binary].active").forEach((b) => b.classList.remove("active"));
  if (state.pending && !state.hasOperand) {
    document.querySelector(`[data-binary="${state.pending.op}"]`)?.classList.add("active");
  }
}

function setExpression(text) {
  els.expression.textContent = text;
}

function setError(text) {
  els.error.textContent = text || "";
}

function renderHistory() {
  els.list.replaceChildren();
  els.empty.hidden = state.history.length > 0;
  for (const calc of [...state.history].reverse()) {
    const li = els.template.content.firstElementChild.cloneNode(true);
    li.querySelector(".history-expr").textContent = describe(calc.operation, calc.operands, calc.angle_unit);
    li.querySelector(".history-result").textContent = describeResult(calc);
    li.querySelector(".history-use").addEventListener("click", () => useValue(calc.result));
    li.querySelector(".history-delete").addEventListener("click", () => deleteCalculation(calc.id));
    els.list.append(li);
  }
}

// ---------- API ----------

async function request(method, path, body) {
  const res = await fetch(API + path, {
    method,
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });
  if (res.status === 204) return null;
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const err = data.error || {};
    throw new Error(ERROR_MESSAGES[err.code] || err.message || `Erro HTTP ${res.status}`);
  }
  return data;
}

// calculate posts a calculation and returns its result, or null on failure.
async function calculate(op, operands) {
  state.busy = true;
  els.display.classList.add("busy");
  try {
    const calc = await request("POST", "/calculations", { operation: op, operands, angle_unit: unit() });
    state.history.push(calc);
    renderHistory();
    setError("");
    return calc.result;
  } catch (e) {
    setError(e.message === "Failed to fetch" ? "Servidor indisponível" : e.message);
    return null;
  } finally {
    state.busy = false;
    els.display.classList.remove("busy");
  }
}

async function loadHistory() {
  try {
    state.history = (await request("GET", "/calculations")).calculations;
    renderHistory();
  } catch (e) {
    setError("Não foi possível carregar o histórico");
  }
}

async function deleteCalculation(id) {
  try {
    await request("DELETE", `/calculations/${id}`);
  } catch (e) {
    setError(e.message);
  }
  state.history = state.history.filter((c) => c.id !== id);
  renderHistory();
}

async function clearHistory() {
  await Promise.allSettled(state.history.map((c) => request("DELETE", `/calculations/${c.id}`)));
  await loadHistory();
}

async function loadDescriptions() {
  try {
    const { operations } = await request("GET", "/operations");
    for (const op of operations) {
      document.querySelectorAll(`[data-unary="${op.name}"], [data-binary="${op.name}"]`)
        .forEach((b) => { b.title = op.description; });
    }
  } catch { /* tooltips are optional */ }
}

// ---------- calculator actions ----------

function useValue(n) {
  state.entry = String(n);
  state.fresh = true;
  state.hasOperand = true;
  setError("");
  render();
}

function inputDigit(d) {
  setError("");
  if (state.fresh) {
    state.entry = d === "." ? "0." : d;
    state.fresh = false;
  } else if (d === "." && state.entry.includes(".")) {
    return;
  } else {
    state.entry = state.entry === "0" && d !== "." ? d : state.entry + d;
  }
  state.hasOperand = true;
  render();
}

async function binary(op) {
  const p = state.pending;
  if (p && state.hasOperand) {
    const y = Number(state.entry);
    const r = await calculate(p.op, [p.x, y]);
    if (r === null) return;
    state.entry = String(r);
  }
  const x = Number(state.entry);
  state.pending = { op, x };
  state.fresh = true;
  state.hasOperand = false;
  const prompts = { root: `raiz de ${operand(x)}, índice:`, log: `log de ${operand(x)}, base:` };
  setExpression(prompts[op] ?? `${operand(x)} ${BINARY_SYMBOLS[op]}`);
  render();
}

async function equals() {
  const p = state.pending;
  if (!p) return;
  const y = Number(state.entry);
  const r = await calculate(p.op, [p.x, y]);
  if (r === null) return;
  setExpression(`${describe(p.op, [p.x, y])} =`);
  state.pending = null;
  useValue(r);
}

async function unary(op) {
  const x = Number(state.entry);
  const r = await calculate(op, [x]);
  if (r === null) return;
  const prefix = state.pending ? `${operand(state.pending.x)} ${BINARY_SYMBOLS[state.pending.op] ?? state.pending.op} ` : "";
  setExpression(`${prefix}${describe(op, [x], unit())} =`);
  useValue(r);
}

const actions = {
  clear() {
    state.pending = null;
    setExpression("");
    actions["clear-entry"]();
  },
  "clear-entry"() {
    setError("");
    useValue(0);
  },
  backspace() {
    if (state.fresh) return;
    state.entry = state.entry.slice(0, -1);
    if (state.entry === "" || state.entry === "-") state.entry = "0";
    render();
  },
  negate() {
    if (state.entry === "0") return;
    state.entry = state.entry.startsWith("-") ? state.entry.slice(1) : "-" + state.entry;
    state.hasOperand = true;
    render();
  },
  equals,
};

const CONSTANTS = { pi: Math.PI, e: Math.E };

function handle(button) {
  if (state.busy) return;
  const { digit, binary: bin, unary: un, action, const: c } = button.dataset;
  if (digit) inputDigit(digit);
  else if (bin) binary(bin);
  else if (un) unary(un);
  else if (c) useValue(CONSTANTS[c]);
  else if (action) actions[action]();
}

// ---------- events ----------

document.querySelectorAll(".keys button").forEach((b) => b.addEventListener("click", () => handle(b)));
$("clear-history").addEventListener("click", clearHistory);

const KEYMAP = {
  "+": '[data-binary="add"]',
  "-": '[data-binary="subtract"]',
  "*": '[data-binary="multiply"]',
  "/": '[data-binary="divide"]',
  "%": '[data-binary="mod"]',
  "^": '[data-binary="power"]',
  ".": '[data-digit="."]',
  ",": '[data-digit="."]',
  "=": '[data-action="equals"]',
  Enter: '[data-action="equals"]',
  Backspace: '[data-action="backspace"]',
  Escape: '[data-action="clear"]',
  Delete: '[data-action="clear-entry"]',
};

document.addEventListener("keydown", (e) => {
  if (e.ctrlKey || e.metaKey || e.altKey) return;
  const selector = /^[0-9]$/.test(e.key) ? `[data-digit="${e.key}"]` : KEYMAP[e.key];
  const button = selector && document.querySelector(selector);
  if (!button) return;
  e.preventDefault();
  button.classList.add("pressed");
  setTimeout(() => button.classList.remove("pressed"), 100);
  handle(button);
});

render();
loadHistory();
loadDescriptions();
