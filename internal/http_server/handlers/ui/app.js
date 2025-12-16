async function apiFetch(path, options) {
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });

  let payload = null;
  const text = await res.text();
  if (text) {
    try {
      payload = JSON.parse(text);
    } catch {
      payload = null;
    }
  }

  if (!res.ok) {
    const msg = payload?.error?.message || `Request failed (${res.status})`;
    const err = new Error(msg);
    err.status = res.status;
    throw err;
  }
  return payload?.data;
}

function setMsg(el, kind, msg) {
  el.classList.remove("ok", "err");
  if (!msg) {
    el.textContent = "";
    return;
  }
  el.textContent = msg;
  el.classList.add(kind);
}

const packsTbody = document.getElementById("packsTbody");
const packsMsg = document.getElementById("packsMsg");
const createForm = document.getElementById("createForm");
const createSize = document.getElementById("createSize");
const resetBtn = document.getElementById("resetBtn");

const calcForm = document.getElementById("calcForm");
const calcQty = document.getElementById("calcQty");
const calcMsg = document.getElementById("calcMsg");
const calcResult = document.getElementById("calcResult");

function renderPacks(packs) {
  packsTbody.innerHTML = "";
  for (const p of packs) {
    const tr = document.createElement("tr");

    const tdId = document.createElement("td");
    tdId.textContent = String(p.id);

    const tdSize = document.createElement("td");
    const input = document.createElement("input");
    input.type = "number";
    input.min = "1";
    input.step = "1";
    input.value = String(p.size);
    input.disabled = true;
    tdSize.appendChild(input);

    const tdActions = document.createElement("td");
    const actions = document.createElement("div");
    actions.className = "row-actions";

    const editBtn = document.createElement("button");
    editBtn.className = "btn btn-secondary";
    editBtn.textContent = "Edit";

    const saveBtn = document.createElement("button");
    saveBtn.className = "btn btn-primary";
    saveBtn.textContent = "Save";
    saveBtn.style.display = "none";

    const cancelBtn = document.createElement("button");
    cancelBtn.className = "btn";
    cancelBtn.textContent = "Cancel";
    cancelBtn.style.display = "none";

    const delBtn = document.createElement("button");
    delBtn.className = "btn btn-danger";
    delBtn.textContent = "Delete";

    let original = p.size;
    function setEditing(on) {
      input.disabled = !on;
      editBtn.style.display = on ? "none" : "";
      saveBtn.style.display = on ? "" : "none";
      cancelBtn.style.display = on ? "" : "none";
      delBtn.disabled = on;
      if (!on) input.value = String(original);
    }

    editBtn.addEventListener("click", (e) => {
      e.preventDefault();
      original = Number(input.value);
      setEditing(true);
      setMsg(packsMsg, "", "");
    });

    cancelBtn.addEventListener("click", (e) => {
      e.preventDefault();
      setEditing(false);
      setMsg(packsMsg, "", "");
    });

    saveBtn.addEventListener("click", async (e) => {
      e.preventDefault();
      const newSize = Number(input.value);
      if (!Number.isInteger(newSize) || newSize <= 0) {
        setMsg(packsMsg, "err", "size must be > 0");
        return;
      }
      try {
        await apiFetch(`/api/packs/${p.id}`, {
          method: "PUT",
          body: JSON.stringify({ size: newSize }),
        });
        setMsg(packsMsg, "ok", "Updated");
        await loadPacks();
      } catch (err) {
        setMsg(packsMsg, "err", err.message);
      } finally {
        setEditing(false);
      }
    });

    delBtn.addEventListener("click", async (e) => {
      e.preventDefault();
      if (!confirm(`Delete pack size ${p.size}?`)) return;
      try {
        await apiFetch(`/api/packs/${p.id}`, { method: "DELETE" });
        setMsg(packsMsg, "ok", "Deleted");
        await loadPacks();
      } catch (err) {
        setMsg(packsMsg, "err", err.message);
      }
    });

    actions.appendChild(editBtn);
    actions.appendChild(saveBtn);
    actions.appendChild(cancelBtn);
    actions.appendChild(delBtn);
    tdActions.appendChild(actions);

    tr.appendChild(tdId);
    tr.appendChild(tdSize);
    tr.appendChild(tdActions);
    packsTbody.appendChild(tr);
  }
}

async function loadPacks() {
  try {
    const data = await apiFetch("/api/packs/", { method: "GET" });
    renderPacks(data.packs || []);
  } catch (err) {
    setMsg(packsMsg, "err", err.message);
  }
}

createForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const size = Number(createSize.value);
  if (!Number.isInteger(size) || size <= 0) {
    setMsg(packsMsg, "err", "size must be > 0");
    return;
  }
  try {
    await apiFetch("/api/packs/", {
      method: "POST",
      body: JSON.stringify({ size }),
    });
    createSize.value = "";
    setMsg(packsMsg, "ok", "Added");
    await loadPacks();
  } catch (err) {
    setMsg(packsMsg, "err", err.message);
  }
});

resetBtn.addEventListener("click", async (e) => {
  e.preventDefault();
  if (!confirm("Reset pack sizes to defaults?")) return;
  try {
    await apiFetch("/api/packs/reset", { method: "POST" });
    setMsg(packsMsg, "ok", "Reset");
    await loadPacks();
  } catch (err) {
    setMsg(packsMsg, "err", err.message);
  }
});

calcForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  calcResult.innerHTML = "";
  const quantity = Number(calcQty.value);
  if (!Number.isInteger(quantity) || quantity <= 0) {
    setMsg(calcMsg, "err", "quantity must be > 0");
    return;
  }
  try {
    const data = await apiFetch("/api/calculate", {
      method: "POST",
      body: JSON.stringify({ quantity }),
    });
    const packs = data.packs || [];
    if (packs.length === 0) {
      setMsg(calcMsg, "ok", "No packs needed");
      return;
    }
    setMsg(calcMsg, "ok", "Calculated");
    for (const p of packs) {
      const li = document.createElement("li");
      li.textContent = `${p.count} × ${p.size}`;
      calcResult.appendChild(li);
    }
  } catch (err) {
    setMsg(calcMsg, "err", err.message);
  }
});

loadPacks();


