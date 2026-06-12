const preview = document.getElementById("preview");
const eventSource = new EventSource("/events");
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);

  if (data.type === "content") {
    console.log("HTML: ", data.html);
    preview.innerHTML = data.html;
  }

  if (data.type === "scroll") {
    console.log("Cursor line: ", data.cursor_line);
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
