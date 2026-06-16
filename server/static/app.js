let lastScrollLine = null;
const preview = document.getElementById("preview");
const eventSource = new EventSource("/events");
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);

  if (data.type === "content") {
    console.log("HTML: ", data.html);
    preview.innerHTML = data.html;

    if (lastScrollLine !== null) {
      scrollToLine(lastScrollLine);
    }
  }

  if (data.type === "scroll") {
    console.log("Cursor line: ", data.cursor_line);
    lastScrollLine = data.cursor_line;
    scrollToLine(data.cursor_line);
  }
};

function scrollToLine(line) {
  const elements = document.querySelectorAll(`[data-source-line="${line}"]`);
  if (elements.length === 0) return;

  let element = elements[elements.length - 1];
  let rect = element.getBoundingClientRect();
  const firstChildLine = element.querySelector("[data-source-line]");
  if (firstChildLine) {
    const childRect = firstChildLine.getBoundingClientRect();
    rect = {
      top: rect.top,
      height: Math.max(1, childRect.top - rect.top),
    };
  }

  const target =
    window.scrollY + rect.top + rect.height / 2 - window.innerHeight / 2;

  window.scrollTo(0, target);
}
