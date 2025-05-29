# 🃏 Tressette - Traditional Istrian Card Game

![Version](https://img.shields.io/badge/version-1.0.0-blue)

A modern web implementation of the classic Italian trick-taking card game **Tressette**, popular in Istria and Mediterranean regions. Play live with 4 of your friends!

## 🚀 Features

- **Multiplayer Mode**: Real-time gameplay with WebSocket support
- **Desktop First**: Mobile maybe in the future
- **Learning Tools**: Rule guide

## 🎮 Game Rules

Tressette is a 40-card trick-taking game where players aim to capture valuable cards through strategic play. Key elements:

- **Card Ranking**: 3 (high) > 2 > A > Re > Cavallo > Fante > 7 > 6 > 5 > 4 (low)
- **Scoring**:
  - Aces = 1 point
  - 3s/2s/face cards = ⅓ point
  - Last trick = 1 point

[Full rules documentation](https://en.wikipedia.org/wiki/Tressette)

## 💻 Installation

Clone repository

```
git clone https://github.com/TeoMatosevic/tressette
```

Install dependencies

```
cd tressette && go mod download
```

Start development server

```
go run cmd/server/main.go
```
