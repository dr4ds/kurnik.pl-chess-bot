# Kurnik.pl chess bot
Bot to play chess on https://kurnik.pl <br>
Requirements: [golang](https://go.dev) and [npm](https://nodejs.org).

### How to compile program?
```sh
cd dashboard
npm install # creates node_modules directory
npm run build # creates build directory
cd ..
go build # creates main program
```

### How to configure and run program?
1. Set login and password in `settings.json` file.
2. Run program
    ```sh
    ./main  # linux
    .\main  # windows
    ```
3. Process will work in background. 
4. You can see dashboard: http://localhost:8080
