package common

import(
  "fmt"
  "net/http"
)

func DisableCors(w *http.ResponseWriter) {
  (*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func PrintJSON(m map[string]interface{}) {
  fmt.Println("JSON:")
  for k, v := range m {
    switch vv := v.(type) {
      case string:
          fmt.Println(k, "is string", vv)
      case float64:
          fmt.Println(k, "is float64", vv)
      case int:
          fmt.Println(k, "is int", vv)
      case []interface{}:
          fmt.Println(k, "is an array:")
          for i, u := range vv {
              fmt.Println(i, u)
          }
      default:
          fmt.Println(k, "is of a type I don't know how to handle")
    }
  }
}

func ParseFormBadRequest(w http.ResponseWriter, r *http.Request) error {
    err := r.ParseForm()
    if err != nil {
      http.Error(w, "Unable to Parse Request", http.StatusBadRequest)
      return err
    }
    return nil
}

