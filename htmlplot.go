/******** Peter Winzell (c), 5/31/24 *********************************************/

package main

import (
	"fmt"
	"net/http"
)

func about_handler(w http.ResponseWriter, r *http.Request) {
	// ABOUT SECTION HTML CODE
	fmt.Fprintf(w, "<title>VISS Curve logging</title>")
	fmt.Fprintf(w, "Curve logging plotter, subscribing to one vss signal, applying curve logging.")
	fmt.Fprintf(w, "<img src='assets/plotter.png' alt='gopher' style='width:1024px;height:800px;'>")
}

func dynamicHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
				<html lang="en">
				<head>
					<meta charset="UTF-8">
					<meta name="viewport" content="width=device-width, initial-scale=1.0">
					<title>Refresh PNG Image</title>
				<script>
						// Function to refresh the image
						function refreshImage() {
							var img = document.getElementById('image');
							img.src = img.src.split('?')[0] + '?' + new Date().getTime();
						}

						// Function to refresh the image every second
						function startRefreshing() {
							setInterval(refreshImage, 500);
						}
				</script>
				</head>
				<body onload="startRefreshing()">
					<img id="image" src="image.png" alt="Image">
				</body>
		</html>`
	fmt.Fprintf(w, html)
}
