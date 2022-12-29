package scrape

import (
	"log"

	"github.com/gocolly/colly/v2"
	"github.com/kuronosu/schema_scraper/pkg/config"
)

func addPaginationOnHTML(c *colly.Collector, link *config.PaginationLink, visitedUrls map[string]bool) {
	counter := 0
	c.OnHTML(link.Selector, func(e *colly.HTMLElement) {
		visitedUrls[e.Request.URL.String()] = true
		pageUrl := link.GetValue(e)
		if _, ok := visitedUrls[pageUrl]; !ok {
			counter++
			if link.Limit == -1 || counter <= link.Limit {
				e.Request.Visit(pageUrl)
			}
		}
	})
}

func ScrapeList(schema *config.PageSchema, url string) config.ParsedLinks {
	c := colly.NewCollector()
	if schema.Cloudflare {
		c.WithTransport(GetCloudFlareRoundTripper())
	}
	results := make(config.ParsedLinks)

	visitedUrls := make(map[string]bool)

	c.OnHTML(schema.List.ContainerSelector, func(e *colly.HTMLElement) {
		results.Extend(config.ParseLinkData(schema.List.SafeCompile(e)))
	})

	addPaginationOnHTML(c, &schema.List.Pagination.Next, visitedUrls)
	addPaginationOnHTML(c, &schema.List.Pagination.Previous, visitedUrls)

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})

	c.Visit(url)
	return results
}

func removeDuplicatesUrls(urls []string) []string {
	keys := make(map[string]interface{})
	list := []string{}
	for _, entry := range urls {
		if _, exist := keys[entry]; !exist {
			keys[entry] = nil
			list = append(list, entry)
		}
	}
	return list
}

func ScrapeListFlat(schema *config.PageSchema, url string) []string {
	c := colly.NewCollector()
	if schema.Cloudflare {
		c.WithTransport(GetCloudFlareRoundTripper())
	}
	results := make([]string, 0)

	visitedUrls := make(map[string]bool)

	c.OnHTML(schema.List.ContainerSelector, func(e *colly.HTMLElement) {
		results = append(results, config.ExtractLinks(schema.List.SafeCompile(e))...)
	})

	addPaginationOnHTML(c, &schema.List.Pagination.Next, visitedUrls)
	addPaginationOnHTML(c, &schema.List.Pagination.Previous, visitedUrls)

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})

	c.Visit(url)
	c.Wait()

	return removeDuplicatesUrls(results)
}
