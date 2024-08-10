# JSON Database in Golang

## Overview

Built a custom JSON database in Golang, inspired by MongoDB. Features include JSON data storage and concurrency handling with mutexes.

## Setup

1. **Initialize Project**: 
   ```bash
   go mod init your_project_name
   ```

2. **Define Structs**: 
   Created `user` and `address` structs.

3. **Implement Functions**:
   - **Create**: Add records.
   - **Read**: Retrieve records.
   - **Write**: Save to files.
   - **Delete**: Remove records.

## Features

- **Data Handling**: Records stored as JSON files.
- **Concurrency**: Mutexes for thread safety.
- **Error Handling**: Checks for file operations and missing resources.

## Usage

1. **Initialize**: Set up the database.
2. **Add Records**: Use the `Write` function.
3. **Retrieve Records**: Use `Read` or `Read All`.
4. **Delete Records**: Use `Delete`.

## Demo

![Database Demo](https://github.com/KamoEllen/Go-Database/blob/main/Demo.gif)
