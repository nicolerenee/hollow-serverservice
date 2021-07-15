package hollow

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	"go.metalkube.net/hollow/internal/db"
)

// Attributes provide the ability to apply namespaced settings to an entity.
// For example hardware could have attributes in the `com.equinixmetal.api` namespace
// that represents equinix metal specific attributes that are stored in the API.
// The namespace is meant to define who owns the schema and values.
type Attributes struct {
	Namespace string          `json:"namespace"`
	Values    json.RawMessage `json:"values"`
}

// AttributeListParams allow you to filter the results based on attributes
type AttributeListParams struct {
	Namespace        string   `form:"namespace" query:"namespace"`
	Keys             []string `form:"keys" query:"keys"`
	EqualValue       string   `form:"equals" query:"equals"`
	LessThanValue    int      `form:"less-than" query:"less-than"`
	GreaterThanValue int      `form:"greater-than" query:"greater-than"`
}

func (a *Attributes) fromDBModel(dbA db.Attributes) error {
	a.Namespace = dbA.Namespace
	a.Values = json.RawMessage(dbA.Values)

	return nil
}

func (a *Attributes) toDBModel() (db.Attributes, error) {
	dbA := db.Attributes{
		Namespace: a.Namespace,
		Values:    datatypes.JSON(a.Values),
	}

	return dbA, nil
}

func convertFromDBAttributes(dbAttrs []db.Attributes) ([]Attributes, error) {
	attrs := []Attributes{}

	for _, dbA := range dbAttrs {
		a := Attributes{}
		if err := a.fromDBModel(dbA); err != nil {
			return nil, err
		}

		attrs = append(attrs, a)
	}

	return attrs, nil
}

func convertToDBAttributes(attrs []Attributes) ([]db.Attributes, error) {
	dbAttrs := []db.Attributes{}

	for _, a := range attrs {
		dbA, err := a.toDBModel()
		if err != nil {
			return nil, err
		}

		dbAttrs = append(dbAttrs, dbA)
	}

	return dbAttrs, nil
}

func encodeAttributesListParams(alp []AttributeListParams, q url.Values) {
	for i, ap := range alp {
		keyPrefix := fmt.Sprintf("attributes_%d_", i)

		q.Set(keyPrefix+"namespace", ap.Namespace)

		for _, k := range ap.Keys {
			q.Add(keyPrefix+"keys", k)
		}

		switch {
		case ap.LessThanValue != 0:
			q.Set(keyPrefix+"less-than", fmt.Sprint(ap.LessThanValue))
		case ap.GreaterThanValue != 0:
			q.Set(keyPrefix+"greater-than", fmt.Sprint(ap.GreaterThanValue))
		default:
			q.Set(keyPrefix+"equals", ap.EqualValue)
		}
	}
}

func parseQueryAttributesListParams(c *gin.Context) ([]AttributeListParams, error) {
	var err error

	alp := []AttributeListParams{}
	i := 0

	for {
		keyPrefix := fmt.Sprintf("attributes_%d_", i)

		ns := c.Query(keyPrefix + "namespace")
		if ns == "" {
			break
		}

		a := AttributeListParams{
			Namespace: ns,
			Keys:      c.QueryArray(keyPrefix + "keys"),
		}

		equals := c.Query(keyPrefix + "equals")
		if equals != "" {
			a.EqualValue = equals
		}

		lt := c.Query(keyPrefix + "less-than")
		if lt != "" {
			a.LessThanValue, err = strconv.Atoi(lt)
			if err != nil {
				return nil, err
			}
		}

		gt := c.Query(keyPrefix + "greater-than")
		if gt != "" {
			a.GreaterThanValue, err = strconv.Atoi(gt)
			if err != nil {
				return nil, err
			}
		}

		alp = append(alp, a)
		i++
	}

	return alp, nil
}
