// --- API Helpers ---

async function api(method, path, body) {
  const opts = {
    method,
    headers: { 'Content-Type': 'application/json' },
    cache: 'no-store',
  };
  if (body !== undefined && body !== null) opts.body = JSON.stringify(body);
  const res = await fetch(path, opts);
  if (res.status === 204) return null;
  let data;
  try {
    data = await res.json();
  } catch {
    throw new Error(`Server error (${res.status})`);
  }
  if (!res.ok) throw new Error(data.error || 'Request failed');
  return data;
}

function showToast(msg, isError = false) {
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.className = 'toast show' + (isError ? ' error' : '');
  setTimeout(() => { t.className = 'toast'; }, 3000);
}

// --- State ---

let allMembers = [];
let allTrips = [];
let balanceData = { balances: [], settlements: [] };

// --- Tab Navigation ---

document.querySelectorAll('.nav-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.nav-btn').forEach(b => b.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
    btn.classList.add('active');
    document.getElementById('tab-' + btn.dataset.tab).classList.add('active');
  });
});

// --- Load Data ---

async function loadAll() {
  try {
    [allMembers, allTrips, balanceData] = await Promise.all([
      api('GET', '/members'),
      api('GET', '/trips'),
      api('GET', '/balances'),
    ]);
    allMembers = allMembers || [];
    allTrips = allTrips || [];
    balanceData = balanceData || { balances: [], settlements: [] };
    renderDashboard();
    renderMembers();
    renderTrips();
  } catch (e) {
    showToast(e.message, true);
  }
}

// --- Dashboard ---

function renderDashboard() {
  const balCards = document.getElementById('balance-cards');
  const settlList = document.getElementById('settlements-list');

  if (!balanceData.balances || balanceData.balances.length === 0) {
    balCards.innerHTML = '<div class="empty-state">No members yet. Add some members to get started.</div>';
    settlList.innerHTML = '';
    return;
  }

  balCards.innerHTML = balanceData.balances.map(b => {
    const val = b.balance;
    const cls = val > 0.01 ? 'positive' : val < -0.01 ? 'negative' : 'zero';
    const sign = val > 0 ? '+' : '';
    return `<div class="card">
      <div class="card-label">${esc(b.member.name)}</div>
      <div class="card-value ${cls}">${sign}${fmt(val)} HUF</div>
    </div>`;
  }).join('');

  if (!balanceData.settlements || balanceData.settlements.length === 0) {
    settlList.innerHTML = '<div class="empty-state">All settled up!</div>';
    return;
  }

  settlList.innerHTML = balanceData.settlements.map(s =>
    `<div class="settlement-row">
      <span class="settlement-from">${esc(s.from_name)}</span>
      <span class="settlement-arrow">&rarr;</span>
      <span class="settlement-to">${esc(s.to_name)}</span>
      <span class="settlement-trips">${(s.trip_names || []).map(n => esc(n)).join(', ')}</span>
      <span class="settlement-amount">${fmt(s.amount)} HUF</span>
    </div>`
  ).join('');
}

// --- Members ---

function renderMembers() {
  const el = document.getElementById('members-list');
  if (allMembers.length === 0) {
    el.innerHTML = '<div class="empty-state">No members yet. Add your first group member.</div>';
    return;
  }
  el.innerHTML = allMembers.map(m => {
    const bal = balanceData.balances?.find(b => b.member.id === m.id);
    const v = bal ? bal.balance : 0;
    const cls = v > 0.01 ? 'positive' : v < -0.01 ? 'negative' : 'zero';
    const sign = v > 0 ? '+' : '';
    return `<div class="card">
      <div class="card-name">${esc(m.name)}</div>
      <div class="card-value ${cls}" style="font-size:1.125rem">${sign}${fmt(v)} HUF</div>
      <div class="card-actions">
        <button class="btn btn-ghost btn-sm" onclick="editMember(${m.id})">Edit</button>
        <button class="btn btn-danger btn-sm" onclick="deleteMember(${m.id})">Delete</button>
      </div>
    </div>`;
  }).join('');
}

function openMemberModal(member) {
  document.getElementById('member-modal-title').textContent = member ? 'Edit Member' : 'Add Member';
  document.getElementById('member-id').value = member ? member.id : '';
  document.getElementById('member-name').value = member ? member.name : '';
  document.getElementById('member-modal').classList.add('open');
  document.getElementById('member-name').focus();
}

function editMember(id) {
  const m = allMembers.find(x => x.id === id);
  if (m) openMemberModal(m);
}

async function saveMember(e) {
  e.preventDefault();
  const id = document.getElementById('member-id').value;
  const name = document.getElementById('member-name').value.trim();
  try {
    if (id) {
      await api('PUT', `/members/${id}`, { name });
      showToast('Member updated');
    } else {
      await api('POST', '/members', { name });
      showToast('Member added');
    }
    closeModal('member-modal');
    await loadAll();
  } catch (e) {
    showToast(e.message, true);
  }
}

async function deleteMember(id) {
  if (!confirm('Delete this member? This will also remove their payments.')) return;
  try {
    await api('DELETE', `/members/${id}`);
    showToast('Member deleted');
    await loadAll();
  } catch (e) {
    showToast(e.message, true);
  }
}

// --- Trips ---

function renderTrips() {
  const el = document.getElementById('trips-list');
  if (!allTrips || allTrips.length === 0) {
    el.innerHTML = '<div class="empty-state">No trips yet. Create your first trip.</div>';
    return;
  }
  el.innerHTML = allTrips.map(t => {
    const dateStr = t.date ? new Date(t.date).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' }) : '';
    const memberNames = (t.members || []).map(m => m.name).join(', ');
    const numMembers = (t.members || []).length;
    const equalShare = numMembers > 0 ? t.total_cost / numMembers : 0;

    const tripAdj = balanceData.trip_adjustments?.[t.id] || {};

    const breakdownRows = (t.members || []).map(m => {
      const paid = (t.payments || [])
        .filter(p => p.member_id === m.id)
        .reduce((s, p) => s + p.amount, 0);
      const rawDelta = paid - equalShare;
      const adj = tripAdj[m.id] || 0;
      const delta = rawDelta + adj;
      const cls = delta > 0.01 ? 'positive' : delta < -0.01 ? 'negative' : '';
      const sign = delta > 0 ? '+' : '';
      return `<tr>
        <td>${esc(m.name)}</td>
        <td>${fmt(paid)}</td>
        <td>${fmt(equalShare)}</td>
        <td class="${cls}" style="font-weight:600">${sign}${fmt(delta)}</td>
        <td>
          ${(t.payments || []).filter(p => p.member_id === m.id).map(p =>
            `<button class="btn-icon" title="Edit payment" onclick="editPayment(${t.id}, ${p.id}, ${p.amount})">&#9998;</button>
             <button class="btn-icon" title="Delete payment" onclick="deletePayment(${t.id}, ${p.id})">&times;</button>`
          ).join('')}
        </td>
      </tr>`;
    }).join('');

    const totalPaid = (t.payments || []).reduce((s, p) => s + p.amount, 0);
    const isComplete = totalPaid >= t.total_cost && t.total_cost > 0;

    return `<div class="trip-card${isComplete ? ' trip-complete' : ''}" id="trip-${t.id}">
      <div class="trip-header" onclick="toggleTrip(${t.id})">
        <div class="trip-title-group">
          <div class="trip-title">${isComplete ? '<span class="trip-check">&#10003;</span> ' : ''}${esc(t.name)}</div>
          <div class="trip-meta">${dateStr} &middot; ${numMembers} members &middot; ${memberNames}</div>
        </div>
        <div class="trip-cost">${fmt(t.total_cost)} HUF</div>
        <span class="trip-chevron">&#9662;</span>
      </div>
      <div class="trip-body">
        <table class="breakdown-table">
          <thead><tr><th>Member</th><th>Paid</th><th>Fair Share</th><th>Balance</th><th></th></tr></thead>
          <tbody>${breakdownRows || '<tr><td colspan="5" style="text-align:center;color:var(--text-tertiary)">No participants</td></tr>'}</tbody>
        </table>
        <div class="trip-actions-bar">
          <button class="btn btn-primary btn-sm" onclick="openPaymentModal(${t.id})">+ Add Payment</button>
          <button class="btn btn-ghost btn-sm" onclick="editTrip(${t.id})">Edit Expense</button>
          <button class="btn btn-danger btn-sm" onclick="deleteTrip(${t.id})">Delete Expense</button>
        </div>
      </div>
    </div>`;
  }).join('');
}

function toggleTrip(id) {
  document.getElementById('trip-' + id).classList.toggle('open');
}

function openTripModal(trip) {
  document.getElementById('trip-modal-title').textContent = trip ? 'Edit Trip' : 'New Trip';
  document.getElementById('trip-id').value = trip ? trip.id : '';
  document.getElementById('trip-name').value = trip ? trip.name : '';
  document.getElementById('trip-cost').value = trip ? trip.total_cost : '';
  document.getElementById('trip-date').value = trip ? trip.date.substring(0, 10) : new Date().toISOString().substring(0, 10);

  const cbs = document.getElementById('trip-members-checkboxes');
  const tripMemberIds = trip ? (trip.members || []).map(m => m.id) : [];
  cbs.innerHTML = allMembers.map(m => {
    const checked = tripMemberIds.includes(m.id) ? 'checked' : '';
    return `<label class="checkbox-chip">
      <input type="checkbox" name="trip-member" value="${m.id}" ${checked}>
      ${esc(m.name)}
    </label>`;
  }).join('');

  if (allMembers.length === 0) {
    cbs.innerHTML = '<span style="color:var(--text-tertiary);font-size:0.8125rem">Add members first</span>';
  }

  document.getElementById('trip-modal').classList.add('open');
}

function editTrip(id) {
  const t = allTrips.find(x => x.id === id);
  if (t) openTripModal(t);
}

async function saveTrip(e) {
  e.preventDefault();
  const id = document.getElementById('trip-id').value;
  const name = document.getElementById('trip-name').value.trim();
  const totalCost = parseFloat(document.getElementById('trip-cost').value);
  const date = document.getElementById('trip-date').value;
  const memberIDs = Array.from(document.querySelectorAll('input[name="trip-member"]:checked'))
    .map(cb => parseInt(cb.value));

  if (memberIDs.length === 0) {
    showToast('Select at least one participant', true);
    return;
  }

  try {
    if (id) {
      await api('PUT', `/trips/${id}`, { name, total_cost: totalCost, date, member_ids: memberIDs });
      showToast('Trip updated');
    } else {
      await api('POST', '/trips', { name, total_cost: totalCost, date, member_ids: memberIDs });
      showToast('Trip created');
    }
    closeModal('trip-modal');
    await loadAll();
  } catch (e) {
    showToast(e.message, true);
  }
}

async function deleteTrip(id) {
  if (!confirm('Delete this trip and all its payments?')) return;
  try {
    await api('DELETE', `/trips/${id}`);
    showToast('Trip deleted');
    await loadAll();
  } catch (e) {
    showToast(e.message, true);
  }
}

// --- Payments ---

function openPaymentModal(tripId, payment) {
  const trip = allTrips.find(t => t.id === tripId);
  document.getElementById('payment-modal-title').textContent = payment ? 'Edit Payment' : 'Add Payment';
  document.getElementById('payment-id').value = payment ? payment.id : '';
  document.getElementById('payment-trip-id').value = tripId;
  document.getElementById('payment-amount').value = payment ? payment.amount : '';

  const memberGroup = document.getElementById('payment-member-group');
  const sel = document.getElementById('payment-member');

  if (payment) {
    memberGroup.style.display = 'none';
  } else {
    memberGroup.style.display = '';
    sel.innerHTML = (trip?.members || []).map(m =>
      `<option value="${m.id}">${esc(m.name)}</option>`
    ).join('');
  }

  document.getElementById('payment-modal').classList.add('open');
  document.getElementById('payment-amount').focus();
}

function editPayment(tripId, paymentId, amount) {
  openPaymentModal(tripId, { id: paymentId, amount });
}

async function savePayment(e) {
  e.preventDefault();
  const paymentId = document.getElementById('payment-id').value;
  const tripId = document.getElementById('payment-trip-id').value;
  const amount = parseFloat(document.getElementById('payment-amount').value);

  try {
    if (paymentId) {
      await api('PUT', `/trips/${tripId}/payments/${paymentId}`, { amount });
      showToast('Payment updated');
    } else {
      const memberId = parseInt(document.getElementById('payment-member').value);
      await api('POST', `/trips/${tripId}/payments`, { member_id: memberId, amount });
      showToast('Payment added');
    }
    closeModal('payment-modal');
    await loadAll();
  } catch (e) {
    showToast(e.message, true);
  }
}

async function deletePayment(tripId, paymentId) {
  if (!confirm('Delete this payment?')) return;
  try {
    await api('DELETE', `/trips/${tripId}/payments/${paymentId}`);
    showToast('Payment deleted');
    await loadAll();
  } catch (e) {
    showToast(e.message, true);
  }
}

// --- Modals ---

function closeModal(id) {
  document.getElementById(id).classList.remove('open');
}

document.addEventListener('keydown', e => {
  if (e.key === 'Escape') {
    document.querySelectorAll('.modal.open').forEach(m => m.classList.remove('open'));
  }
});

// --- Helpers ---

function esc(str) {
  const d = document.createElement('div');
  d.textContent = str;
  return d.innerHTML;
}

function fmt(n) {
  return Number.isInteger(n) ? n.toString() : n.toFixed(2);
}

// --- Init ---
loadAll();
