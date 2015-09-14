package main

import "net/http"
import "fmt"
import "time"
import "encoding/base64"
import "bytes"
import "io"
import "strconv"

import "localhost/vibrant"
import "image"
import _ "image/jpeg"
import _ "image/png"

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%v %v\n", time.Now(), r)
	w.Header().Set("Content-Type", "text/html")

    maxColors := 65535

	fmt.Fprintf(w, `<!doctype html><html><head><meta charset='utf-8'><title>Go Vibrant!</title><style>
*{box-sizing: border-box;margin: 0;}
body {
    font-family: sans-serif;
    max-width: 800px;
    margin: auto;
}
h3 { color: #800; }
hr { margin: 1em 0; border: 0; border-top: 1px ridge;}
h1, form { display: inline-block; }
form { margin-left: 12%% }
button { background: #9cf; color: #fff; border: 0; padding: .5em 1em; border-radius: 3px; text-transform: uppercase; letter-spacing: .5px; box-shadow: 0 0 5px #ccc; transition: all 125ms cubic-bezier(.8,0,.2,1); cursor: pointer; z-index: 1; position: relative } button:hover { box-shadow: none; } button:active, button:focus { box-shadow: 0 0 0 100vmax rgba(153,204,255,1) }input[type="text"]{margin-right: 1em}</style></head><body><h1>choose an image:</h1><form action='/' method='post' enctype='multipart/form-data'><input type='file' name='test' accept='image/*'><input type="text" size="4" maxlength="10" name="maxColors" title="maxColors (default: %d)"><button type='submit' name='vibrant' value='q'>Go Vibrant!</button></form>`, maxColors)
	defer func() {
		fmt.Fprintf(w, "</body></html>")
	}()

	if r.FormValue("vibrant") == "" {
		//w.WriteHeader(http.StatusOK)
		return
	}

	fmt.Fprintf(w, "<hr>\n")

	file, header, err := r.FormFile("test")
	if err != nil {
		//w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "<h3>Error: 400 Bad Request</h3>")
		return
	}

	switch header.Header["Content-Type"][0] {
	case "image/jpeg":
	case "image/png":
	default:
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "<h3>Error: JPG/PNG only plz.</h3>")
			return
		}
	}

	img, _, err := image.Decode(file)
	file.Seek(0, 0)
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, file)
	file.Close()
	datauri := fmt.Sprintf("data:%s;base64,%s", header.Header["Content-Type"][0], base64.StdEncoding.EncodeToString(buf.Bytes()))
	if err != nil {
		//w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "<h3>Error: %s</h3>", err)
		return
	}

    if max := r.FormValue("maxColors"); max != "" {
        n,err := strconv.Atoi(max)
        if err != nil {
            fmt.Fprintf(w, "<h3>Error: %s</h3>", err)
        }
        maxColors = n
    }

	start := time.Now()
	bitmap := vibrant.NewBitmap(img)
	palette, err := vibrant.Generate(bitmap, maxColors)
	//palette, err := vibrant.NewPalette(bitmap)
	benchmark := time.Since(start)
	if err != nil {
		//w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "<h3>Error: %s</h3>", err)
		return
	}

	fmt.Fprintf(w, `<style>
figure {
    box-shadow: 0 0 10px #ccc;
}
img { max-width: 100%%; display: block; }
figcaption {
    display: flex;
    flex-wrap: wrap;
}
figcaption div {
    flex: 1 1 16.6667%%;
    padding: 3vw 0;
    text-align: center;
    display: inline-block;
}
textarea { margin: 1em auto; width: 100%%; display: block; height: 12em; resize: none; border: 0; }
h2 { text-align: center }`)
	stylesheet := ""
	if palette.VibrantSwatch != nil {
		stylesheet = fmt.Sprintf("%s%s", stylesheet, palette.VibrantSwatch)
		vendorPrefixingIsAWESOME := fmt.Sprintf("{ background-color: %s; color: %s; }", palette.VibrantSwatch.RGBHex(), palette.VibrantSwatch.TitleTextColor())
		fmt.Fprintf(w, "::selection %s\n::-moz-selection %s\n::-webkit-selection %s", vendorPrefixingIsAWESOME, vendorPrefixingIsAWESOME, vendorPrefixingIsAWESOME)
	}
	if palette.DarkVibrantSwatch != nil {
		stylesheet = fmt.Sprintf("%s%s", stylesheet, palette.DarkVibrantSwatch)
	}
	if palette.LightVibrantSwatch != nil {
		stylesheet = fmt.Sprintf("%s%s", stylesheet, palette.LightVibrantSwatch)
	}
	if palette.MutedSwatch != nil {
		stylesheet = fmt.Sprintf("%s%s", stylesheet, palette.MutedSwatch)
	}
	if palette.DarkMutedSwatch != nil {
		stylesheet = fmt.Sprintf("%s%s", stylesheet, palette.DarkMutedSwatch)
	}
	if palette.LightMutedSwatch != nil {
		stylesheet = fmt.Sprintf("%s%s", stylesheet, palette.LightMutedSwatch)
	}

	fmt.Fprintf(w, "%s</style>", stylesheet)
	fmt.Fprintf(w, "<figure><img src='%s'>", datauri)
	fmt.Fprintf(w, `<figcaption>
            <div class="vibrant">Vibrant</div>
            <div class="lightvibrant">LightVibrant</div>
            <div class="darkvibrant">DarkVibrant</div>
            <div class="muted">Muted</div>
            <div class="lightmuted">LightMuted</div>
            <div class="darkmuted">DarkMuted</div>
        </figcaption>
    </figure>`)
	fmt.Fprintf(w, "<textarea readonly onclick='this.select()'>%s</textarea>", stylesheet)
	fmt.Fprintf(w, "<h2>%v</h2>", benchmark)
}

func main() {
	http.HandleFunc("/", index)
	fmt.Println("Listening on 0.0.0.0:8080...")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		fmt.Println("Error:", err)
	}
}
