document.addEventListener("DOMContentLoaded", function() {
  const form = document.getElementById("analyzeForm");
  const loader = document.getElementById("loader");
  const cancelButton = document.getElementById("cancelButton");
  const logArea = document.getElementById("logArea");
  let xhr;
  let eventSource;

  form.addEventListener("submit", function(event) {
    event.preventDefault();
    logArea.innerHTML = ""; // Clear previous logs
    logMessage("Submitting the form...");
    loader.style.display = "block";
    cancelButton.style.display = "block";

    // Start SSE connection
    eventSource = new EventSource('/sse');
    eventSource.onmessage = function(event) {
      logMessage(event.data);
    };

    xhr = new XMLHttpRequest();
    xhr.open("POST", form.action, true);

    xhr.onreadystatechange = function() {
      if (xhr.readyState === 4) {
        loader.style.display = "none";
        cancelButton.style.display = "none";
        eventSource.close(); // Close the SSE connection
        if (xhr.status === 200) {
          logMessage("Analysis completed successfully.");
          document.open();
          document.write(xhr.responseText);
          document.close();
        } else {
          logMessage("Error: " + xhr.status + " " + xhr.statusText);
        }
      }
    };

    xhr.onerror = function() {
      loader.style.display = "none";
      cancelButton.style.display = "none";
      eventSource.close(); // Close the SSE connection
      logMessage("Request failed.");
    };

    xhr.onabort = function() {
      loader.style.display = "none";
      cancelButton.style.display = "none";
      eventSource.close(); // Close the SSE connection
      logMessage("Request canceled.");
    };

    xhr.ontimeout = function() {
      loader.style.display = "none";
      cancelButton.style.display = "none";
      eventSource.close(); // Close the SSE connection
      logMessage("Request timed out.");
    };

    const formData = new FormData(form);
    xhr.send(formData);
  });

  cancelButton.addEventListener("click", function() {
    if (xhr) {
      xhr.abort();
      loader.style.display = "none";
      cancelButton.style.display = "none";
      eventSource.close(); // Close the SSE connection
      logMessage("Request canceled.");
    }
  });

  function logMessage(message) {
    const logEntry = document.createElement("div");
    logEntry.textContent = message;
    logArea.appendChild(logEntry);
    logArea.scrollTop = logArea.scrollHeight;
  }
});
