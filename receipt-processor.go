package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "math/rand"
    "math"
    "time"
    "strconv"
    "strings"
    "github.com/gorilla/mux"
)

// Set up a Receipt struct that takes JSON values, and a Points value as well
type Receipt struct {
  Retailer string `json:"retailer"`
  PurchaseDate string `json:"purchaseDate"`
  PurchaseTime string `json:"purchaseTime"`
  Items []Item `json:"items"`
  Total string `json:"total"`
  Points int
}

// Set up an Item struct which is referenced by Receipt as an array of Items
// Also taking JSON values
type Item struct {
  ShortDescription string `json:"shortDescription"`
  Price string `json:"price"`
}

// Set up a struct for making a JSON object with the receipt id
type Id struct {
  Id string `json:"id"`
}

// Set up a struct for making a JSON object storing the total points on a receipt
type Points struct {
  Points int `json:"points"`
}

// Declare a var map for storing processed receipts
var receipts map[string]Receipt

// The main function
func main() {

    // Create a receipts map for storing receipts that have been processed
    receipts = make(map[string]Receipt)

    // Set up a new router, r
    r := mux.NewRouter()

    // Handle incoming POST requests on /receipts/process
    r.HandleFunc("/receipts/process", func(w http.ResponseWriter, r *http.Request) {
      // Set up a receipt var
      var receipt Receipt
      // Decode the POSTed JSON object and store the values in receipt
      json.NewDecoder(r.Body).Decode(&receipt)
      // Process the receipt and store returned id value in var id
      var id string = processReceipt(receipt)
      // If there was an error during processing
      if id == "Error During Processing" {
        fmt.Fprintf(w, "Error Encountered During Processing of Receipt Data." +
                       " Receipt not stored. Expected a parseable number," +
                       " time, or date.")
      // Otherwise, if no error,
      } else {
        // Create the id struct with the value of id
        receiptId := Id { Id: id }
        // Make JSON version of the struct, and write to http.ResponseWriter
        json.NewEncoder(w).Encode(receiptId)
      }
    })

    // Handle incoming GET requests on /receipts/{id}/points
    r.HandleFunc("/receipts/{id}/points", func(w http.ResponseWriter, r *http.Request) {
      // Grab the id from the requested route
      vars := mux.Vars(r)
      id := vars["id"]
      // If the id exists in the receipts map...
      if _, ok := receipts[id]; ok {
        // Create the Points struct for the given receipt id
        receiptPoints := Points { Points: receipts[id].Points }
        // Make a JSON version of the Points struct and write to http.ResponseWriter
        json.NewEncoder(w).Encode(receiptPoints)
      } else {
        fmt.Fprintf(w, "Requested receipt id %s was not found in the stored" +
          " processed receipts", id)
      }
    })

    // Listen on port 80 on localhost
    http.ListenAndServe(":80", r)
}

// The function for processing the receipt coming in via the /receipts/process route
// Which returns a string that is the id for the receipt being processed
func processReceipt( receipt Receipt ) string {
    // First, tally up the points and store it in the receipt struct
    receipt.Points = tallyPoints(receipt)
    // If the tallyPoints function encountered an error during parsing strings,
    // That is, if it returned -1,
    if receipt.Points == -1 {
      // Return error message to main function
      return "Error During Processing"
    }
    // Second, create a random string of characters and numbers and dashes
    receiptId := String(32)
    // If, by some off chance we already have the same id in use in the map,
    if _, ok := receipts[receiptId]; ok {
      // Make a new random string to replace it
      receiptId = String(32)
    }
    // Third, store the receipt in the receipts map with the id as key to the map
    receipts[receiptId] = receipt
    // Finally, return the id to the main function
    return receiptId
}

// The function for tallying points, returns an integer number of Points
func tallyPoints( receipt Receipt ) int {

    // Set totalPoints to zero for starters
    var totalPoints int = 0


    // Rule 1: One point for every alphanumeric character in the retailer name

    // Grab the retailer string from the receipt being processed
    var retailer string = receipt.Retailer
    // Iterate over all the characters in the retailer's name
    for n := 0; n < len(retailer); n++ {
      // c is the character in question
      c := retailer[n]

      // If c is between A and Z, increment totalPoints by one
      if c >= 'A' && c <= 'Z' {
          totalPoints++
      }
      // If c is between a and z, increment totalPoints by one
      if c >= 'a' && c <= 'z' {
          totalPoints++
      }
      // If c is between 0 and 9, increment totalPoints by one
      if c >= '0' && c <= '9' {
          totalPoints++
      }
        // If c is anything else, it is ignored
    }


    // Rule 2: 50 Points if the total is a round dollar amount,
    // & Rule 3: 25 Points if the total is a multiple of 0.25

    // Grab the total from the receipt being processed, make it a float64
    totalPrice, err := strconv.ParseFloat(receipt.Total, 64)
    // If there's no error from the ParseFloat function
    if err == nil {
      // If the float64 and integer values are identical, add 50 Points
      // to totalPoints because this means it's a whole dollar amount
      if totalPrice == float64(int(totalPrice)) {
          totalPoints += 50
      }
      // If the modulus of the total divided by 0.25 is zero, this means
      // it is a multiple of 0.25, and thus we add 25 Points to totalPoints
      if math.Mod(totalPrice, 0.25) == 0.0 {
          totalPoints += 25
      }
      // If there was an error,
    } else {
      return -1
      // Return -1 to the calling function indicating an error during processing
    }


    // Rule 4: 5 Points for every two items on the receipt

    // Using integer division, the number of items divided by two is the number
    // of pairs not including any odd items, and then multiply by 5 Points
    // and add this multiple to totalPoints
    totalPoints += ( len(receipt.Items) / 2 ) * 5


    // Rule 5: If the trimmed length of item description is a multiple of 3,
    // multiply th price by 0.2 and round up to the nearest integer, add this
    // to the total number of points earned

    // Iterate through all the items in the receipt being processed
    for n := 0; n < len(receipt.Items) ; n++ {
      // Create an item to work with, taking the nth one from the array
      var item = receipt.Items[n]
      // Trim the extra whitespace off the short description of the item
      trimmedDescription := strings.TrimSpace(item.ShortDescription)
      // If the length of the trimmed description is evenly divisible by 3
      // (it's a multiple of 3)
      if len(trimmedDescription) % 3 == 0 {
        // Grab the item price and convert to a float64
        price, err := strconv.ParseFloat(item.Price, 64)
        // If no error was thrown on the ParseFloat function
        if err == nil {
          // Take the ceiling (round up to nearest integer) of the Price
          // multiplied by 0.2 and convert to an int and add to totalPoints
          totalPoints += int(math.Ceil(price * 0.2))
          // If there was an error,
        } else {
          // Return -1 to the calling function, indicating an error during processing
          return -1
        }
      }
    }


    // Rule 6: 6 Points if the day of the purchase is odd

    // Split the date on hyphen and make it into an array
    var date = strings.Split(receipt.PurchaseDate, "-")
    // If the date is correctly formatted (3 numbers split by two hyphens)
    if len(date) == 3 {
      // Convert the day value to an integer
      day, err := strconv.Atoi(date[2])
      // If there's no error coming from the Atoi function
      if err == nil {
        // Test if the modulus of day and 2 is not zero (meaning day is odd)
        if day % 2 != 0 {
          // If so, add 6 Points to totalPoints
          totalPoints += 6
        }
        // If there was an error,
      } else {
        // Return -1 to the calling function, indicating an error during processing
        return -1
      }
    // If the date was not correctly formatted,
    } else {
      // Return -1 to the calling function, indicating an error during processing
      return -1
    }


    // Rule 7: 10 Points if the time of purchase is after 2pm and before 4pm

    // Split the time on semicolon and make it into an array
    var time = strings.Split(receipt.PurchaseTime, ":")
    // If the time is correctly formatted (two numbers split by a semicolon)
    if len(time) == 2 {
      // Convert the hour value to an integer
      hour, err := strconv.Atoi(time[0])
      // If there's no error coming from the Atoi function
      if err == nil {
        // Test if the hour is between 2PM and 4PM (1400 hours and 1600 hours)
        if hour >= 14 && hour <= 16 {
          // If so, add 10 points to totalPoints
          totalPoints += 10
        }
        // If there was an error,
      } else {
        // Return -1 to the calling function, indicating an error during processing
        return -1
      }
    // If the time is not correctly formatted,
    } else {
      // Return -1 to the calling function, indicating an error during processing
      return -1
    }

    // Return the number of total points for the receipt being processed
    return totalPoints
}

// define the set of characters to use for random id
const charset = "abcdefghijklmnopqrstuvwxyz" +
  "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"

// Set up random seed using the present time
var seededRand *rand.Rand = rand.New(
  rand.NewSource(time.Now().UnixNano()))

// Make a string of a particular length, and from a particular char set
func StringWithCharset(length int, charset string) string {
  b := make([]byte, length)
  for i := range b {
    b[i] = charset[seededRand.Intn(len(charset))]
  }
  return string(b)
}

// Make a string of a particular length, using the charset defined above
func String(length int) string {
  return StringWithCharset(length, charset)
}
