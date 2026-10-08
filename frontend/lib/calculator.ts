/**
 * Pure logic for the Step 1 "Quick Calculator".
 *
 * The expression is kept as a plain string made of digits, ".", and the
 * operators + - * /  (e.g. "1200+350*2"). Nothing here uses eval().
 *
 * Rules enforced while typing:
 *  - a number can have at most MAX_DECIMALS decimals and MAX_INT_DIGITS integer digits
 *  - only one "." per number, no leading zeros ("007" -> "7")
 *  - consecutive operators replace each other ("5+" then "*" -> "5*")
 *  - a leading "-" is allowed (negative start), other leading operators are ignored
 */

export type Operator = "+" | "-" | "*" | "/";

export type CalcResult =
  | { ok: true; value: number }
  | { ok: false; error: "incomplete" | "divide-by-zero" | "too-large" };

export const MAX_DECIMALS = 2;
export const MAX_INT_DIGITS = 9;
/** Largest value that may be applied back to the amount field. */
export const MAX_AMOUNT = 999_999_999;

const OPERATORS: readonly string[] = ["+", "-", "*", "/"];

export const isOperator = (c: string | undefined): c is Operator =>
  c !== undefined && OPERATORS.includes(c);

/** Index where the number currently being typed starts. */
function currentNumberStart(expr: string): number {
  for (let i = expr.length - 1; i >= 0; i--) {
    // an operator at index 0 is a sign ("-5"), so the number starts right after it
    if (isOperator(expr[i])) return i + 1;
  }
  return 0;
}

const currentNumber = (expr: string) => expr.slice(currentNumberStart(expr));

export function appendDigit(expr: string, digit: string): string {
  const seg = currentNumber(expr);
  const [intPart, decPart] = seg.split(".");

  if (decPart !== undefined) {
    if (decPart.length >= MAX_DECIMALS) return expr;
    return expr + digit;
  }

  if (seg === "0") {
    // replace the lone zero instead of building "05"
    return digit === "0" ? expr : expr.slice(0, -1) + digit;
  }

  if (intPart.length >= MAX_INT_DIGITS) return expr;
  return expr + digit;
}

export function appendDecimal(expr: string): string {
  const seg = currentNumber(expr);
  if (seg.includes(".")) return expr;
  return seg === "" ? expr + "0." : expr + ".";
}

export function appendOperator(expr: string, op: Operator): string {
  if (expr === "") return op === "-" ? "-" : expr;
  if (expr === "-") return expr;

  const last = expr[expr.length - 1];
  if (isOperator(last)) return expr.slice(0, -1) + op;
  if (last === ".") return expr.slice(0, -1) + op;
  return expr + op;
}

export function backspace(expr: string): string {
  return expr.slice(0, -1);
}

/**
 * Evaluate with normal precedence (* and / before + and -).
 * A trailing operator is ignored so the live result keeps showing while the
 * user is mid-typing ("1200+" still shows 1,200).
 */
export function evaluate(expr: string): CalcResult {
  let s = expr;
  while (s.length > 0 && isOperator(s[s.length - 1])) s = s.slice(0, -1);
  if (s === "" || s === "-" || s === ".") return { ok: false, error: "incomplete" };

  const nums: number[] = [];
  const ops: Operator[] = [];
  let cur = "";

  for (let i = 0; i < s.length; i++) {
    const c = s[i];
    if (isOperator(c) && !(i === 0 && c === "-")) {
      nums.push(parseFloat(cur));
      ops.push(c);
      cur = "";
    } else {
      cur += c;
    }
  }
  nums.push(parseFloat(cur));

  // pass 1: * and /
  const terms: number[] = [nums[0]];
  const addOps: Operator[] = [];
  for (let i = 0; i < ops.length; i++) {
    const n = nums[i + 1];
    if (ops[i] === "*") {
      terms[terms.length - 1] *= n;
    } else if (ops[i] === "/") {
      if (n === 0) return { ok: false, error: "divide-by-zero" };
      terms[terms.length - 1] /= n;
    } else {
      addOps.push(ops[i]);
      terms.push(n);
    }
  }

  // pass 2: + and -
  let total = terms[0];
  addOps.forEach((op, i) => {
    total = op === "+" ? total + terms[i + 1] : total - terms[i + 1];
  });

  // round to paise so 0.1 + 0.2 is 0.3, and 10 / 3 is 3.33
  const rounded = Math.round(total * 100) / 100;
  if (!Number.isFinite(rounded) || Math.abs(rounded) > MAX_AMOUNT) {
    return { ok: false, error: "too-large" };
  }
  return { ok: true, value: rounded };
}

/** Number -> string suitable for TransactionFormData.amount ("1630", "12.5", "99.99"). */
export function toAmountString(value: number): string {
  return String(Number(value.toFixed(2)));
}

/** "1234567.5" -> "12,34,567.5" (Indian digit grouping). Keeps a trailing "." while typing. */
export function formatIndian(val: string): string {
  if (!val) return "0";
  const negative = val.startsWith("-");
  const unsigned = negative ? val.slice(1) : val;
  const [intPart, decPart] = unsigned.split(".");
  const digits = intPart === "" ? "0" : intPart;

  const lastThree = digits.slice(-3);
  const rest = digits.slice(0, -3);
  const grouped = rest ? rest.replace(/\B(?=(\d{2})+(?!\d))/g, ",") + "," + lastThree : lastThree;

  return (negative ? "-" : "") + grouped + (decPart !== undefined ? "." + decPart : "");
}

const OP_SYMBOL: Record<Operator, string> = { "+": "+", "-": "−", "*": "×", "/": "÷" };

/** "1200+350*2" -> "1,200 + 350 × 2" for display. */
export function formatExpression(expr: string): string {
  let out = "";
  let cur = "";
  for (let i = 0; i < expr.length; i++) {
    const c = expr[i];
    if (isOperator(c) && !(i === 0 && c === "-")) {
      out += formatIndian(cur) + ` ${OP_SYMBOL[c]} `;
      cur = "";
    } else {
      cur += c;
    }
  }
  return out + (cur === "" ? "" : formatIndian(cur));
}
