package model

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Places are where you go to lunch
type Places struct {
	Session      *mgo.Session
	DatabaseName string
}

// NewPlaces creates a Places struct
func NewPlaces(session *mgo.Session, databaseName string) *Places {
	places := &Places{
		Session:      session,
		DatabaseName: databaseName,
	}

	places.ensurePlacesIndex()

	return places
}

func (places Places) ensurePlacesIndex() {
	session := places.Session.Copy()
	defer session.Close()

	c := session.DB(places.DatabaseName).C("places")

	nameIndex := mgo.Index{
		Key:        []string{"name", "teamid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err := c.EnsureIndex(nameIndex)
	if err != nil {
		log.Fatal(err)
	}

	teamIndex := mgo.Index{
		Key:        []string{"teamid"},
		Unique:     false,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(teamIndex)
	if err != nil {
		log.Fatal(err)
	}
}

// AllPlaces returns all places for this team
func (places Places) AllPlaces(teamID string) ([]Place, error) {
	session := places.Session.Copy()
	defer session.Close()

	c := session.DB(places.DatabaseName).C("places")

	var thePlaces []Place
	err := c.Find(bson.M{"teamid": teamID}).All(&thePlaces)
	if err != nil {
		log.Println("Failed to get all places: ", err)
		return nil, fmt.Errorf("Database error")
	}

	return thePlaces, nil
}

// FindByID returns a single place with team ID & place ID
func (places Places) FindByID(teamID string, id string) (Place, error) {
	session := places.Session.Copy()
	defer session.Close()

	c := session.DB(places.DatabaseName).C("places")

	var place Place
	err := c.Find(bson.M{"teamid": teamID, "_id": bson.ObjectIdHex(id)}).One(&place)
	if err != nil {
		log.Printf("Failed to find place %v error: %v\n", id, err)
		return Place{}, fmt.Errorf("Database error")
	}

	return place, nil
}

func onlyProposablePlaces(vs []Place) []Place {
	vsf := make([]Place, 0)
	for _, v := range vs {
		if v.LastSkipped.Before(time.Now().Add(-time.Hour*6)) &&
			v.LastVisited.Before(time.Now().Add(-time.Hour*72)) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func sortProposablePlaces(vs []Place) []Place {
	// TODO: weight places so we prefer places we haven't been to / skipped recently
	for i := range vs {
		j := rand.Intn(i + 1)
		vs[i], vs[j] = vs[j], vs[i]
	}
	return vs
}

// ProposePlace picks a place to go to for lunch based on a proprietary algorithm (basically randomness)
func (places Places) ProposePlace(teamID string) (Place, error) {
	session := places.Session.Copy()
	defer session.Close()

	c := session.DB(places.DatabaseName).C("places")

	var allPlaces []Place
	err := c.Find(bson.M{"teamid": teamID}).All(&allPlaces)
	if err != nil {
		log.Println("Failed to get all places: ", err)
		return Place{}, fmt.Errorf("Database error")
	}

	allPlaces = onlyProposablePlaces(allPlaces)

	if len(allPlaces) == 0 {
		log.Printf("We have nowhere to go :(")
		return Place{}, fmt.Errorf("There are no places that haven't been skipped or visited recently")
	}

	allPlaces = sortProposablePlaces(allPlaces)

	log.Printf("We have %v places!\n", len(allPlaces))

	return allPlaces[0], nil
}

// AddPlace adds a new place
func (places Places) AddPlace(place Place) (string, error) {
	session := places.Session.Copy()
	defer session.Close()

	c := session.DB(places.DatabaseName).C("places")

	log.Printf("Adding %v in %v", place.Name, place.TeamID)

	err := c.Insert(place)
	if err != nil {
		if mgo.IsDup(err) {
			return "", fmt.Errorf("A place with this name already exists")
		}

		log.Println("Failed to insert place: ", err)
		return "", fmt.Errorf("Database error")
	}

	return place.ID.Hex(), nil
}

// func placeByID(s *mgo.Session) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		session := s.Copy()
// 		defer session.Close()

// 		id := pat.Param(r, "id")
// 		c := session.DB("lunch").C("places")

// 		var place Place
// 		err := c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&place)
// 		if err != nil {
// 			support.ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
// 			log.Printf("Failed to find place %v error: %v\n", id, err)
// 			return
// 		}

// 		if place.ID == "" {
// 			support.ErrorWithJSON(w, "Place not found", http.StatusNotFound)
// 		}

// 		respBody, err := json.MarshalIndent(place, "", "  ")
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		support.ResponseWithJSON(w, respBody, http.StatusOK)
// 	}
// }

// UpdatePlace updates a given place in the database
func (places Places) UpdatePlace(teamID string, id string, updates interface{}) error {
	session := places.Session.Copy()
	defer session.Close()

	c := session.DB(places.DatabaseName).C("places")

	err := c.Update(bson.M{
		"_id":    bson.ObjectIdHex(id),
		"teamid": teamID,
	}, bson.M{
		"$set": &updates,
	})
	if err != nil {
		if mgo.IsDup(err) {
			return fmt.Errorf("A place with that name already exists")
		}
		switch err {
		case mgo.ErrNotFound:
			return fmt.Errorf("Place not found")
		default:
			log.Println("Failed to update place: ", err)
			return fmt.Errorf("Database error")
		}
	}

	return nil
}

// VisitPlace updates the database to record that user has accepted our lunch suggestion.
func (places Places) VisitPlace(teamID string, id string) (Place, error) {
	session := places.Session.Copy()
	defer session.Close()

	c := session.DB(places.DatabaseName).C("places")

	var place Place
	err := c.Find(bson.M{"teamid": teamID, "_id": bson.ObjectIdHex(id)}).One(&place)
	if err != nil {
		log.Printf("Failed to find place %v error: %v\n", id, err)
		return Place{}, fmt.Errorf("Database error")
	}

	place.LastVisited = time.Now()
	place.VisitCount++

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)}, &place)
	if err != nil {
		if mgo.IsDup(err) {
			return Place{}, fmt.Errorf("A place with that name already exists")
		}
		switch err {
		case mgo.ErrNotFound:
			return Place{}, fmt.Errorf("Place not found")
		default:
			log.Println("Failed to update place: ", err)
			return Place{}, fmt.Errorf("Database error")
		}
	}

	return place, nil
}

// SkipPlace records that the a user has decided the algorithm is wrong and that this is not the place to go to today.
func (places Places) SkipPlace(teamID string, id string) error {
	session := places.Session.Copy()
	defer session.Close()

	c := session.DB(places.DatabaseName).C("places")

	var place Place
	err := c.Find(bson.M{"teamid": teamID, "_id": bson.ObjectIdHex(id)}).One(&place)
	if err != nil {
		log.Printf("Failed to find place %v error: %v\n", id, err)
		return fmt.Errorf("Database error")
	}

	place.LastSkipped = time.Now()
	place.SkipCount++

	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)}, &place)
	if err != nil {
		if mgo.IsDup(err) {
			return fmt.Errorf("A place with this name already exists")
		}
		switch err {
		case mgo.ErrNotFound:
			return fmt.Errorf("Place not found")
		default:
			log.Println("Failed to update place: ", err)
			return fmt.Errorf("Database error")
		}
	}

	return nil
}

// DeletePlace removes a single place from the database
func (places Places) DeletePlace(teamID string, id string) error {
	session := places.Session.Copy()
	defer session.Close()

	c := session.DB(places.DatabaseName).C("places")

	query := bson.M{"teamid": teamID, "_id": bson.ObjectIdHex(id)}
	err := c.Remove(query)
	if err != nil {
		switch err {
		case mgo.ErrNotFound:
			return fmt.Errorf("Place not found")
		default:
			log.Println("Failed to delete place: ", err)
			return fmt.Errorf("Database error")
		}
	}

	return nil
}
