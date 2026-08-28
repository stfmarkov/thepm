document.addEventListener("click", function (event) {
	var btn = event.target.closest("[data-copy-target]");
	if (!btn) {
		return;
	}
	var el = document.getElementById(btn.getAttribute("data-copy-target"));
	if (!el) {
		return;
	}
	var text = el.value || el.textContent || "";
	if (navigator.clipboard && navigator.clipboard.writeText) {
		navigator.clipboard.writeText(text);
	}
});
