# book-crud2

A simple CRUD application for managing books, written in Go.

## Prerequisites

*   [Go](https://golang.org/)
*   [MySQL](https://www.mysql.com/)

## Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/nallyhoo/book-crud2.git
    cd book-crud2
    ```

2.  **Install dependencies:**
    ```bash
    go mod tidy
    ```

3.  **Set up the database:**
    *   Make sure you have MySQL running.
    *   Create a database named `bookdb`.
    *   The application will automatically create the `books` table.

4.  **Update the database connection string:**
    *   Open `main.go` and update the `dsn` variable with your MySQL connection details:
        ```go
        dsn := "YOUR_USERNAME:YOUR_PASSWORD@tcp(127.0.0.1:3306)/bookdb?parseTime=true"
        ```

## Usage

1.  **Run the application:**
    ```bash
    go run main.go
    ```

2.  Open your web browser and navigate to [http://localhost:8080](http://localhost:8080).
