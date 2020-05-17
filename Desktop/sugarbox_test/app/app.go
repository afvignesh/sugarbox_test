package app

import(
	"net/http"
	u "../utils"
	c "../constants"
	m "../mongo_wrapper"
	"../models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/gorilla/mux"
	"encoding/json"
	"context"
)

var UserName string = "francis1"

var FetchUserActivity = func(w http.ResponseWriter, r *http.Request) {

	userActivity := models.UserActivity{}
	param, ok := r.URL.Query()["user_name"] // would be fetched from JWT Token in normal cases ...
	
	if ok && len(param[0]) > 0 {
		UserName = param[0]
	}

	condition := bson.D{{"username", UserName}}

	err := m.MongoFindOneRecord(c.ACTIVITY_COLLECTION, condition, &userActivity)
	
	if err != nil{
		response := u.Message(false, err.Error())
		u.Respond(w, response)
		return
	}

	results := []models.MovieInfo{}

	moviesList := userActivity.Movies

	condition = bson.D{{"name", bson.D{{"$in", moviesList}} }}

	cur, err := m.MongoFindMultipleRecord(c.MOVIES_COLLECTION, condition)

	for cur.Next(context.TODO()) {
		var elem models.MovieInfo
		ratingArr := make([]models.Rating,0)
		commentArr := make([]models.Comment,0)
        err = cur.Decode(&elem)
        if err != nil {
            response := u.Message(false, err.Error())
			u.Respond(w, response)
			return
		}

		for _, rating := range elem.Ratings{
			if rating.UserName == UserName {
				ratingArr = append(ratingArr, rating)
			}
		}

		elem.Ratings = ratingArr

		for _, comment := range elem.Comments{
			if comment.UserName == UserName {
				commentArr = append(commentArr, comment)
			}
		}

		elem.Comments = commentArr
    
        results = append(results, elem)
    }
    
    if err = cur.Err(); err != nil {
		response := u.Message(false, err.Error())
		u.Respond(w, response)
		return
    }
    
    // Close the cursor once finished
    cur.Close(context.TODO())

	response := u.Message(true, "Data found Successfully.")
	response["data"] = results
	u.Respond(w, response)
}

// Fetching movies based on search result ... 
var FindAllMovies = func(w http.ResponseWriter, r *http.Request) {


	param, ok := r.URL.Query()["searchText"]
	
	condition := bson.D{} // if no search text then return entire list of movies ...
	if ok && len(param[0]) > 0 {
		condition = bson.D{{"name", primitive.Regex{Pattern: param[0], Options: "i"}}} // case insensitive search ...
	}

	cur, err := m.MongoFindMultipleRecord(c.MOVIES_COLLECTION, condition)
	
	if err != nil {
		response := u.Message(false, err.Error())
		u.Respond(w, response)
		return
	}

	results := make([]models.MovieInfo, 0)
    for cur.Next(context.TODO()) {
		var elem models.MovieInfo
        err = cur.Decode(&elem)
        if err != nil {
            response := u.Message(false, err.Error())
			u.Respond(w, response)
			return
		}
		elem.RatingCount = len(elem.Ratings)
		elem.AvgRating = 0

		for _, rating := range elem.Ratings{
			elem.AvgRating += float64(rating.Rating)/float64(elem.RatingCount)
		}
    
        results = append(results, elem)
    }
    
    if err = cur.Err(); err != nil {
		response := u.Message(false, err.Error())
		u.Respond(w, response)
		return
    }
    
    // Close the cursor once finished
    cur.Close(context.TODO())

	response := u.Message(true, "Movie found Successfully.")
	response["data"] = results
	u.Respond(w, response)

}

func checkIfMovieExists(moviename string, moviesArr []string) (bool){
	ok := false

	for _, movie := range moviesArr{
		if moviename == movie {
			ok = true
			break
		}
	}

	return ok
}

func updateUserActivity(username string, movieName string) (error){
	condition :=  bson.D{{"username", username}}
	userActivity := models.UserActivity{}

	userActivity.UserName = username
	userActivity.Movies = append(userActivity.Movies, movieName)
	err := m.MongoFindOneRecord(c.ACTIVITY_COLLECTION, condition, &userActivity)
	if err != nil {
		if err.Error() == "mongo: no documents in result"{
			err = insertActivity(userActivity)
			return err
		} else {
			return err
		}
	}

	ok := checkIfMovieExists(movieName, userActivity.Movies)
	if !ok {
		update := bson.M{"$push":bson.M{"movies":movieName}}
		_, err := m.MongoUpdateOneRecord(c.ACTIVITY_COLLECTION, condition, update)
		return err
	}

	return nil
}

func insertActivity(userActivity models.UserActivity) (error){
	_, err := m.MongoInsertOneRecord(c.ACTIVITY_COLLECTION, userActivity)
	return err
}

func isRatingPresent(addedRating models.Rating, RatingArr []models.Rating) ([]models.Rating, bool){
	ok := false

	for i := 0; i < len(RatingArr); i++ {
		if RatingArr[i].UserName == addedRating.UserName {
			RatingArr[i].Rating = addedRating.Rating
			ok = true
		}
	}
	return RatingArr, ok
}

// Adding ratings to the given movie ...
var AddUserRating = func(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	movieName := params["mname"]
	condition :=  bson.D{{"name", movieName}}

	movieInfo := models.MovieInfo{}

	err := m.MongoFindOneRecord(c.MOVIES_COLLECTION, condition, &movieInfo) // check wether movie is present or not ...

	if err != nil {
		response := u.Message(false, err.Error())
		u.Respond(w, response)
		return
	}

	addedRating := models.Rating{}
	err = json.NewDecoder(r.Body).Decode(&addedRating)
	
	if err != nil{
		response := u.Message(false, err.Error())
		u.Respond(w, response)
		return
	}

	ratingArr, ok := isRatingPresent(addedRating, movieInfo.Ratings)

	user_name := addedRating.UserName
	
	update := bson.M{"$push":bson.M{"ratings":addedRating}}

	if ok {
		update = bson.M{"$set":bson.M{"ratings": ratingArr}}
	}

	result, err := m.MongoUpdateOneRecord(c.MOVIES_COLLECTION, condition, update)
	
	if err != nil {
		if err.Error() == "mongo: no documents in result"{
			response := u.Message(false, "Invalid Id")
			u.Respond(w, response)
			return
		}
		response := u.Message(false, "Connection error. Please retry")
		u.Respond(w, response)
		return
	}

	err = updateUserActivity(user_name, movieName)

	if err != nil {
		response := u.Message(false, err.Error())
		u.Respond(w, response)
		return
	}
	
	responseText := "Matched "+ string(result.MatchedCount) +" documents and updated "+ string(result.ModifiedCount)+ "documents."
	response := u.Message(true, responseText)
	u.Respond(w, response)

}

// Adding comments to the given movie ...
var AddUserComment = func(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	movieName := params["mname"]
	condition :=  bson.D{{"name", movieName}}

	addedComment := models.Comment{}
	err := json.NewDecoder(r.Body).Decode(&addedComment)

	movieInfo := models.MovieInfo{}

	err = m.MongoFindOneRecord(c.MOVIES_COLLECTION, condition, &movieInfo) // check wether movie is present or not ...

	if err != nil {
		response := u.Message(false, err.Error())
		u.Respond(w, response)
		return
	}
	
	update := bson.M{"$push":bson.M{"comments":addedComment}}

	result, err := m.MongoUpdateOneRecord(c.MOVIES_COLLECTION, condition, update)
	
	if err != nil {
		if err.Error() == "mongo: no documents in result"{
			response := u.Message(false, "Invalid Id")
			u.Respond(w, response)
			return
		}
		response := u.Message(false, "Connection error. Please retry")
		u.Respond(w, response)
		return
	}
	user_name := addedComment.UserName
	err = updateUserActivity(user_name, movieName)

	if err != nil {
		response := u.Message(false, err.Error())
		u.Respond(w, response)
		return
	}
	
	responseText := "Matched "+ string(result.MatchedCount) +" documents and updated "+ string(result.ModifiedCount)+ "documents."
	response := u.Message(true, responseText)
	u.Respond(w, response)
}