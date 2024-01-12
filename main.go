package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"


	"github.com/chromedp/chromedp"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/html"
)

func main() {
	var url, class, id string
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "class",
				Usage: "Css class of desired SVGs",
			},
			&cli.StringFlag{
				Name:  "id",
				Usage: "Css ID of desired SVGs",
			},
		},
		Action: func(cCtx *cli.Context) error {
			// var url string
			if cCtx.NArg() < 1 {
				return fmt.Errorf("You must provide a URL")
			} else {
				url = cCtx.Args().Get(0)
			}
			if cCtx.String("class") != "" {
				class = cCtx.String("class")
			}
			if cCtx.String("id") != "" {
				id = cCtx.String("id")
			}
			svgSources, err := fetchSVGsByClassOrID(url, class, id)
			if err != nil {
				log.Fatal(err)
			}
			for i, svgSource := range svgSources {
				filename := fmt.Sprintf("svg_%d.svg", i+1)
				if err := saveSVGToFile(filename, svgSource); err != nil {
					log.Printf("Error saving SVG %d: %v", i+1, err)
					return err
				}
			}

			return nil

		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func fetchSVGsByClassOrID(url, className string, idName string) ([]string, error) {
	// Create a new context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Navigate to the specified URL
	if err := chromedp.Run(ctx, chromedp.Navigate(url)); err != nil {
		return nil, err
	}

	// Wait for JavaScript to execute (adjust the duration as needed)
	// time.Sleep(2 * time.Second)

	// Capture the HTML content after JavaScript execution
	var htmlContent string
	if err := chromedp.Run(ctx, chromedp.OuterHTML("html", &htmlContent)); err != nil {
		return nil, err
	}

	// Parse the HTML document
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	// Find all SVG elements with the specified class
	var svgSources []string
	var findSVG func(*html.Node)
	findSVG = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "svg" {
			// Check if the SVG has the specified class
			if idName != "" {
				for _, attr := range n.Attr {
					if attr.Key == "id" && strings.Contains(attr.Val, idName) {
						// Serialize the HTML of the SVG element
						var sb strings.Builder
						html.Render(&sb, n)
						svgSources = append(svgSources, sb.String())
						break
					}
				}
			}
			if className != "" {
				for _, attr := range n.Attr {
					if attr.Key == "class" && strings.Contains(attr.Val, className) {
						var sb strings.Builder
						html.Render(&sb, n)
						svgSources = append(svgSources, sb.String())
						break
					}
				}
			} else if idName == "" && className == "" {
				// Just get all SVGs
				var sb strings.Builder
				html.Render(&sb, n)
				svgSources = append(svgSources, sb.String())
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findSVG(c)
		}
	}

	findSVG(doc)

	return svgSources, nil
}

func saveSVGToFile(filename, svgContent string) error {
	return ioutil.WriteFile(filename, []byte(svgContent), 0644)
}
