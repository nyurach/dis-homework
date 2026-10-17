const $ = (id) => document.getElementById(id);
let catalog = [],
  selected,
  table,
  editing,
  revision = 0;
const references = {
  student_id: { path: "students/students", label: "full_name" },
  teacher_id: { path: "teachers/teachers", label: "full_name" },
  course_id: { path: "courses/courses", label: "title" },
};
let relatedRows = {};

function displayValue(column, value) {
  const reference = references[column];
  if (!reference) return value ?? "—";
  const record = (relatedRows[column] || []).find(
    (row) => String(row.id) === String(value),
  );
  return record
    ? `${record[reference.label]} (#${value})`
    : `Missing record #${value}`;
}

async function api(path, options) {
  const r = await fetch("/api/" + path, options);
  const data = await r.json();
  if (!r.ok) throw Error(data.error || "Request failed");
  return data;
}
function node(tag, text) {
  const el = document.createElement(tag);
  el.textContent = text;
  return el;
}
function path() {
  return selected.name + "/" + table.name;
}
function choose(db) {
  selected = db;
  document.querySelectorAll(".database").forEach((b, i) => {
    b.classList.toggle("active", catalog[i] === db);
    b.setAttribute("aria-pressed", catalog[i] === db);
  });
  $("title").textContent = db.title;
  table = db.tables[0];
  load();
}
async function load() {
  const current = ++revision;
  $("notice").textContent = "Loading…";
  $("body").replaceChildren();
  $("head").replaceChildren();
  try {
    const activeTable = table;
    const rows = await api(path());
    const entries = await Promise.all(
      activeTable.columns
        .filter((column) => references[column])
        .map(async (column) => {
          return [column, await api(references[column].path)];
        }),
    );
    if (current !== revision) return;
    relatedRows = Object.fromEntries(entries);
    const tr = document.createElement("tr");
    ["id", ...table.columns, "actions"].forEach((c) =>
      tr.append(node("th", c.replaceAll("_", " "))),
    );
    $("head").replaceChildren(tr);
    rows.forEach((row) => {
      const tr = document.createElement("tr");
      ["id", ...table.columns].forEach((c) =>
        tr.append(node("td", displayValue(c, row[c]))),
      );
      const actions = document.createElement("td");
      actions.className = "actions";
      const edit = node("button", "Edit");
      edit.onclick = () => openEditor(row);
      const del = node("button", "Delete");
      del.onclick = async () => {
        if (!confirm("Delete record #" + row.id + "?")) return;
        del.disabled = true;
        try {
          await api(path() + "/" + row.id, { method: "DELETE" });
          load();
        } catch (e) {
          $("notice").textContent = e.message;
          del.disabled = false;
        }
      };
      actions.append(edit, del);
      tr.append(actions);
      $("body").append(tr);
    });
    $("notice").textContent = rows.length + " records";
  } catch (e) {
    if (current === revision) $("notice").textContent = e.message;
  }
}
function openEditor(row) {
  editing = row;
  $("formTitle").textContent =
    (row ? "Edit" : "Add") + " " + table.name.replaceAll("_", " ");
  $("formError").textContent = "";
  $("fields").replaceChildren();
  table.columns.forEach((c) => {
    const label = node(
      "label",
      c.replaceAll("_", " ") + (c === "grade" ? " (0–100)" : ""),
    );
    const input = document.createElement(references[c] ? "select" : "input");
    if (references[c]) {
      const placeholder = node("option", "Select a record");
      placeholder.value = "";
      input.append(placeholder);
      (relatedRows[c] || []).forEach((record) => {
        const option = node(
          "option",
          `${record[references[c].label]} (#${record.id})`,
        );
        option.value = record.id;
        input.append(option);
      });
    }
    input.name = c;
    input.value = row?.[c] ?? "";
    input.required = true;
    if (input.tagName === "INPUT") {
      if (c.endsWith("_id") || ["grade_level", "grade"].includes(c)) {
        input.type = "number";
        input.step = c === "grade" ? "0.01" : "1";
        input.min = c === "grade" ? "0" : "1";
        if (c === "grade") input.max = "100";
        if (c === "grade_level") input.max = "12";
      }
      if (c === "address") input.type = "url";
    }
    label.append(input);
    $("fields").append(label);
  });
  $("editor").showModal();
}
$("form").onsubmit = async (e) => {
  e.preventDefault();
  $("save").disabled = true;
  const data = Object.fromEntries(new FormData(e.target));
  try {
    await api(path() + (editing ? "/" + editing.id : ""), {
      method: editing ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });
    $("editor").close();
    await load();
  } catch (e) {
    $("formError").textContent = e.message;
  } finally {
    $("save").disabled = false;
  }
};
$("cancel").onclick = () => $("editor").close();
$("add").onclick = () => {
  if (table) openEditor(null);
};
api("databases")
  .then((data) => {
    catalog = data;
    data.forEach((db, i) => {
      const b = node("button", "");
      b.className = "database";
      b.append(node("small", "DATABASE 0" + (i + 1)), node("strong", db.title));
      b.onclick = () => choose(db);
      $("databases").append(b);
    });
    choose(data[0]);
  })
  .catch((e) => {
    $("notice").textContent = e.message;
    $("title").textContent = "Connection failed";
  });
