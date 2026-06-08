const preview = document.getElementById("preview");
const eventSource = new EventSource("/events");
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);

  if (data.type === "content") {
    preview.innerHTML = data.html;
  }

  if (data.type === "scroll") {
    scrollToLine(data.cursor_line);
  }
};

function scrollToLine(line) {
  const element = document.querySelector(`[data-source-line="${line}"]`);
  if (element) {
    element.scrollIntoView({
      behavior: "instant",
      block: "center",
    });
  }
}
