package main

func getJoke(jokeId int) (Joke, error) {

	res := Joke{}

	var id int
	var joke string

	err := db.QueryRow(`Select id, joke from Jokes where id = $1`, jokeId).Scan(&id, &joke)
	if err == nil {
		res = Joke{ID: id, Joke: joke}
	}

	return res, err
}

func allJokes() ([]Joke, error) {

	jokes := []Joke{}

	rows, err := db.Query(`SELECT id, joke from jokedata order by id`)

	if err == nil {
		for rows.Next() {
			var id int
			var joke string

			err = rows.Scan(&id, &joke)
			if err == nil {
				currentJoke := Joke{ID: id, Joke: joke}
				jokes = append(jokes, currentJoke)
			} else {
				return jokes, err
			}
		}
	}

	return jokes, err
}

func createJoke(joke string) (int, error) {

	var jokeID int
	err := db.QueryRow(`INSERT INTO jokedata(joke) VALUES($1) RETURNING id`, joke).Scan(&jokeID)

	if err != nil {
		return 0, err
	}

	return jokeID, err
}

func updateJoke(id int, joke string) (int, error) {

	res, err := db.Exec(`UPDATE jokedata set joke=$1 where id=$2 RETURNING id`, joke, id)
	if err != nil {
		return 0, err
	}

	rowsUpdated, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsUpdated), err
}

func deleteJoke(jokeID int) (int, error) {

	res, err := db.Exec(`DELETE from jokedata where id = $1`, jokeID)
	if err != nil {
		return 0, err
	}

	rowsDeleted, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsDeleted), nil
}
