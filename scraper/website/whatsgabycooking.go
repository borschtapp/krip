package website

import (
	"github.com/borschtapp/krip/model"
)

func ScrapeWhatsGabyCooking(data *model.DataInput, r *model.Recipe) error {
	r.Publisher.Name = "What's Gaby Cooking"
	return nil
}
