package mongo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGenerateNumberFromPrimitiveId(t *testing.T) {
	id, _ := primitive.ObjectIDFromHex("66af422044f3b3de8e272a2a")
	number := generateIntegerFromObjectId(id)

	assert.Equal(t, 2566698, number, id.Hex())
}
func TestGenerateTranslitName(t *testing.T) {
	id, _ := primitive.ObjectIDFromHex("66af422044f3b3de8e272a2a")

	cases := []struct {
		Name         string
		Value        string
		ExceptedName string
	}{
		{
			Name:         "regular translition",
			Value:        "Книга по здоровому питанию",
			ExceptedName: "kniga-po-zdorovomu-pitaniju-2566698",
		},
		{
			Name:         "with english letters",
			Value:        "Storytelling: книга о квадроберах",
			ExceptedName: "storytelling-kniga-o-kvadroberah-2566698",
		},
		{
			Name:         "different symbols",
			Value:        "Здесь !\";№%:?*()много сим%;%;:волов",
			ExceptedName: "zdes-mnogo-simvolov-2566698",
		},
		{
			Name:         "with dash saved",
			Value:        "Бизнес-проекция. Инфографика",
			ExceptedName: "biznes-proekcija-infografika-2566698",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			name := GenerateTranslitName(c.Value, id)
			assert.Equal(t, c.ExceptedName, name)
		})
	}
}
