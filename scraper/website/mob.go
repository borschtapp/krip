package website

import (
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

func ScrapeMob(data *model.DataInput, r *model.Recipe) error {
	r.Publisher.Url = utils.BaseUrl(r.Url)
	return nil
}
