document.addEventListener("DOMContentLoaded", function() {
  const form = document.getElementById("analyzeForm");
  const loaderContainer = document.querySelector(".loader-container");
  const loader = document.getElementById("loader");
  const cancelButton = document.getElementById("cancelButton");
  const logArea = document.getElementById("logArea");
  let xhr;
  let eventSource;

  form.addEventListener("submit", function(event) {
      event.preventDefault();
      logArea.innerHTML = "";
      logMessage("Submitting the form...");
      loaderContainer.style.display = "flex";

      eventSource = new EventSource('/sse');
      eventSource.onmessage = function(event) {
          logMessage(event.data);
      };

      xhr = new XMLHttpRequest();
      xhr.open("POST", form.action, true);

      xhr.onreadystatechange = function() {
          if (xhr.readyState === 4) {
              loaderContainer.style.display = "none";
              eventSource.close();
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
          loaderContainer.style.display = "none";
          eventSource.close();
          logMessage("Request failed.");
      };

      xhr.onabort = function() {
          loaderContainer.style.display = "none";
          eventSource.close();
          logMessage("Request canceled.");
      };

      xhr.ontimeout = function() {
          loaderContainer.style.display = "none";
          eventSource.close();
          logMessage("Request timed out.");
      };

      xhr.timeout = 30000; // Set timeout

      const formData = new FormData(form);
      xhr.send(formData);
  });

  cancelButton.addEventListener("click", function() {
      if (xhr) {
          xhr.abort();
          loaderContainer.style.display = "none";
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
