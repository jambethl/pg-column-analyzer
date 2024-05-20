# PostgreSQL Column Analyzer

PostgreSQL Column Analyzer is a CLI tool written in Go which iterates through your PostgreSQL tables and produces a CSV file detailing the column order of each table, reporting how much space is wasted due to inefficient column definition order, helping you optimize your database schemas for ideal performance.

## Features
- Connects to your PostgreSQL instance, iterating through its tables in the specified schema
- Analyzes the column order and calculates the space wasted due to inefficient column order definitions
- Generates a CSV report with the optimal column order and space-saving suggestions.
