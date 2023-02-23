package website

import (
	"github.com/borschtapp/krip/model"
)

func ScrapeArchanasKitchen(data *model.DataInput, r *model.Recipe) error {
	r.Publisher.Name = "Archana's Kitchen"
	return nil
}
