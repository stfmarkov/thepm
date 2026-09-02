document.addEventListener("click", function (event) {
	var copyBtn = event.target.closest("[data-copy-target]");
	if (copyBtn) {
		var el = document.getElementById(copyBtn.getAttribute("data-copy-target"));
		if (el) {
			var text = el.value || el.textContent || "";
			if (navigator.clipboard && navigator.clipboard.writeText) {
				navigator.clipboard.writeText(text);
			}
		}
		return;
	}

	var addBtn = event.target.closest("[data-add-link], [data-add-note]");
	if (addBtn) {
		openRowComposer(addBtn.closest(".section"));
		return;
	}

	var editBtn = event.target.closest("[data-edit-link], [data-edit-note]");
	if (editBtn) {
		var item = editBtn.closest(".link-item, .note-item");
		if (!item) {
			return;
		}
		item.classList.add("is-editing");
		focusRowField(item);
		return;
	}

	var cancelBtn = event.target.closest("[data-cancel-edit]");
	if (cancelBtn) {
		event.preventDefault();
		closeRowEdit(cancelBtn.closest(".link-item, .note-item"));
		return;
	}

	var sortBtn = event.target.closest("[data-feedback-sort]");
	if (sortBtn) {
		feedbackSort = sortBtn.getAttribute("data-feedback-sort") || "newest";
		applyFeedbackSort(sortBtn.closest("#project-feedback"));
	}
});

document.addEventListener("keydown", function (event) {
	if (event.key !== "Escape") {
		return;
	}
	closeRowEdit(event.target.closest(".link-item.is-editing, .note-item.is-editing"));
});

function rowConfig(section) {
	if (!section) {
		return null;
	}
	if (section.id === "project-notes") {
		return {
			newRow: "#note-new",
			empty: "[data-notes-empty]",
			add: "[data-add-note]",
			item: ".note-item:not(.is-new)",
			focus: "[name='body']",
		};
	}
	if (section.id === "project-links") {
		return {
			newRow: "#link-new",
			empty: "[data-links-empty]",
			add: "[data-add-link]",
			item: ".link-item:not(.is-new)",
			focus: "[name='label']",
		};
	}
	return null;
}

function focusRowField(item) {
	var field = item.querySelector(".link-edit [name='label'], .note-edit [name='body']");
	if (field) {
		field.focus();
	}
}

function openRowComposer(section) {
	var cfg = rowConfig(section);
	if (!cfg) {
		return;
	}
	var row = section.querySelector(cfg.newRow);
	if (!row) {
		return;
	}
	var empty = section.querySelector(cfg.empty);
	if (empty) {
		empty.hidden = true;
	}
	var addBtn = section.querySelector(cfg.add);
	if (addBtn) {
		addBtn.hidden = true;
	}
	row.hidden = false;
	var field = row.querySelector(cfg.focus);
	if (field) {
		field.focus();
	}
}

function closeRowEdit(item) {
	if (!item) {
		return;
	}
	var form = item.querySelector("form.link-edit, form.note-edit");
	if (form) {
		form.reset();
	}
	if (item.classList.contains("is-new")) {
		item.hidden = true;
		var section = item.closest(".section");
		var cfg = rowConfig(section);
		if (!cfg) {
			return;
		}
		var addBtn = section.querySelector(cfg.add);
		if (addBtn) {
			addBtn.hidden = false;
		}
		var empty = section.querySelector(cfg.empty);
		var hasRows = section.querySelector(cfg.item);
		if (empty) {
			empty.hidden = !!hasRows;
		}
		return;
	}
	item.classList.remove("is-editing");
}

var feedbackSort = "newest";

document.addEventListener("htmx:afterSwap", function (event) {
	var root = event.detail && event.detail.target;
	if (!root || !root.querySelector) {
		return;
	}
	var section = root.id === "project-feedback" ? root : root.querySelector("#project-feedback");
	if (section) {
		applyFeedbackSort(section);
	}
});

function parseFeedbackRating(value) {
	if (!value) {
		return null;
	}
	var n = Number(value);
	if (!isFinite(n)) {
		return null;
	}
	return n;
}

function compareFeedbackItems(a, b) {
	var ra = parseFeedbackRating(a.getAttribute("data-rating"));
	var rb = parseFeedbackRating(b.getAttribute("data-rating"));
	var ta = Number(a.getAttribute("data-received")) || 0;
	var tb = Number(b.getAttribute("data-received")) || 0;
	if (feedbackSort === "oldest") {
		return ta - tb;
	}
	if (feedbackSort === "highest" || feedbackSort === "lowest") {
		if (ra === null && rb === null) {
			return tb - ta;
		}
		if (ra === null) {
			return 1;
		}
		if (rb === null) {
			return -1;
		}
		if (ra !== rb) {
			return feedbackSort === "lowest" ? ra - rb : rb - ra;
		}
		return tb - ta;
	}
	return tb - ta;
}

function applyFeedbackSort(section) {
	if (!section) {
		return;
	}
	var list = section.querySelector("[data-feedback-list]");
	if (list) {
		var items = Array.prototype.slice.call(list.querySelectorAll(".feedback-item"));
		items.sort(compareFeedbackItems);
		items.forEach(function (el) {
			list.appendChild(el);
		});
	}
	var buttons = section.querySelectorAll("[data-feedback-sort]");
	Array.prototype.forEach.call(buttons, function (btn) {
		var on = btn.getAttribute("data-feedback-sort") === feedbackSort;
		btn.classList.toggle("is-active", on);
		btn.setAttribute("aria-pressed", on ? "true" : "false");
	});
}
