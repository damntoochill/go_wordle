package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
)

var client *redis.Client

func Init(redisUrl string) error {
	ctx := context.Background()
	if redisUrl == "" {
		return fmt.Errorf("empty redis url")
	}
	client = redis.NewClient(&redis.Options{Addr: redisUrl, Password: "", DB: 0})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return err
	}

	return nil
}

func addToRedis(words []string) error {
	ctx := context.Background()
	for i := range words {
		result := client.ZAdd(ctx, "wordle", &redis.Z{Score: 1, Member: words[i]})
		if result.Err() != nil {
			return result.Err()
		}
	}
	return nil
}

func validGuess(guess string) bool {
	ctx := context.Background()
	result := client.ZScore(ctx, "wordle", guess)
	if result.Val() != 1 {
		return false
	}
	return true
}

func readFile() ([]string, error) {
	file, err := os.Open("wordslist.txt")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	wordsList, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	words := strings.Split(string(wordsList), "\n")

	return words, nil
}

func main() {

	const rules = `Wordle Rules:
		1. You have 5 chances to guess the word correctly.
		2. When you enter a guess the correct letters at correct position turns green.
		3. The correct letters at wrong positions turn yellow
		4. The wrong letters turn gray
		5. Based on the clues you get in previous guesses you can make your new guess.`

	redisUrl := os.Getenv("REDIS_URL")

	err := Init(redisUrl)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rules)

	var found bool

	words, err := readFile()
	if err != nil {
		log.Fatalln(err)
	}

	err = addToRedis(words)
	if err != nil {
		log.Fatal(err)
	}

	k := rand.Intn(len(words))
	word := words[k]

	lettersMap := make(map[int]string, 0)

	for i := range word {
		lettersMap[i] = string(word[i])
	}

	in := bufio.NewReader(os.Stdin)

	fmt.Println("Type your guess and press enter")
	for !found {
		fmt.Print("-->")
		guess, _ := in.ReadString('\n')
		guess = strings.ReplaceAll(guess, "\n", "")
		guess = strings.ReplaceAll(guess, "\r", "")

		if len(guess) != 5 {
			fmt.Println("Enter only 5 Letter words")
			continue
		}

		if !validGuess(guess) {
			fmt.Println("Enter a valid word")
			continue
		}

		for i := range guess {
			if guess[i] == word[i] {
				fmt.Print("\033[32m" + string(guess[i]) + "\033[0m")
			} else if strings.ContainsAny(word, string(guess[i])) {
				if string(guess[i]) == lettersMap[i] {
					fmt.Print("\033[32m" + string(guess[i]) + "\033[0m")
				}
				fmt.Print("\033[33m" + string(guess[i]) + "\033[0m")
			} else {
				fmt.Print("\033[37m" + string(guess[i]) + "\033[0m")
			}
		}
		fmt.Println()

		if guess == word {
			fmt.Println("Congratulations you won!")
			found = true
		}
	}
}
