# Web Page Analyzer
This is a web application built with Go that analyzes a web page given its URL. The analysis includes:

* HTML version of the document.
* Page title.
* Number of headings of different levels.
* Number of internal and external links.
* Number of inaccessible links.
* Presence of a login form.
* Real-time logging of the analysis process.


## How to run:
```
git clone https://github.com/yourusername/web-page-analyzer.git
cd web-page-analyzer
go mod tidy
go run main.go
```

To view the application, open your browser and go to:
```
http://localhost:8080
```

## Possible Improvements
- Enhance Error Handling: Improve error messages to be more user-friendly.

- UI Enhancements: Use a modern frontend framework like React or Vue.js for a better user experience.

- Caching: Use an in-memory cache like go-cache or a distributed cache like Redis.

- Implement more comprehensive accessibility checks.