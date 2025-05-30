let ws
let myPlayerId = null // Store the player's ID received from the server
let myPlayerName = null // Store the player's name
let teamsInfo = null // Store teams info if needed
let trickCards = [] // Store cards played in the current trick
let handCards = [] // Store cards in hand

let playDisabled = false // Flag to disable play card action
let afterTrick = false // Flag to indicate if the trick has ended
let roundOver = false // Flag to indicate if the round has ended
let roundOverPayload = null // Store the payload for round over
let gameOver = false // Flag to indicate if the game is over
let canDeclare = false // Flag to indicate if the player can declare

let declarations = [{
    type: "napola",
    suit: "Bastoni",
    rank: "",
    image: "images/cards/Bastoni_1.png",
}, {
    type: "napola",
    suit: "Kope",
    rank: "",
    image: "images/cards/Kope_1.png",
}, {
    type: "napola",
    suit: "Denari",
    rank: "",
    image: "images/cards/Denari_1.png",
}, {
    type: "napola",
    suit: "Spade",
    rank: "",
    image: "images/cards/Spade_1.png",
}, {
    type: "three_or_four_of_kind",
    suit: "",
    rank: "1",
    image: "images/cards/Bastoni_1.png",
}, {
    type: "three_or_four_of_kind",
    suit: "",
    rank: "2",
    image: "images/cards/Bastoni_2.png",
}, {
    type: "three_or_four_of_kind",
    suit: "",
    rank: "3",
    image: "images/cards/Bastoni_3.png",
}]


const rules_text = `<h2>Rules</h2>
    <p>Tressette is a traditionl Italian and Istrian card game played with a 40-card deck. 
    The game is played by four players in two teams of two players each.</p>
    <p>The objective of the game is to score points by winning tricks and declaring special combinations of cards.</p>
    <p>Each player is dealt 10 cards. Cards are divided into four suits: Bastoni, Kope, Denari, and Spade. </p>
    <p>The strongest card in each suit is 3, followed by 2, 1, 13, 12, 11, 7, 6, 5, and 4.</p>
    <p>Ace or 1 is the card that values the most (1 point). 2, 3, 13, 12, 11 are worth 1/3 of a point and 7, 6, 5, and 4 are worth 0 points.</p>
    <p>Bastoni look like clubs, Kope look like cups, Denari look like coins, and Spade look like swords.</p>
    <p>Whoever wins the first trick gets an additional point.</p>
    <p>The first player to play a card leads the trick. The next player must follow suit if possible. Only cards of the same suit matter when calculating the winner of the trick.
    If a player cannot follow suit, they can play any card but it will never win the trick.</p>
    <p>The player who wins the trick leads the next trick. The game continues until all cards have been played.</p>
    <p>Players also have the option to declare special combinations of cards for additional points before their first turn.</p>
    <p>This means everyone knows that the a certain player has certain cards in their hand, but than that player's team gets additional points.</p>
    <p>Declarations are:</p>
    <p>Napola means that the player has 3, 2 and 1 of the same suit. The player gets 3 points for this declaration.</p>
    <p>Three of a kind means that the player has 3 cards of the same rank (only cards with ranks 3, 2, 1 are valid for this declaration). The player gets 3 points for this declaration.</p>
    <p>Four of a kind means that the player has 4 cards of the same rank (only cards with ranks 3, 2, 1 are valid for this declaration). The player gets 4 points for this declaration.</p>
    <p>The game is supposed to be played in silence.</p>
    <p>Game ends when one of the teams reaches the points goal.</p>
    <p>Have fun!</p>
    <p>For more information about the game, visit the <a href="https://en.wikipedia.org/wiki/Tressette" target="_blank">Wikipedia page</a>.</p>
`


// DOM Elements
const statusMessage = document.getElementById("status-message")
const playerHandDiv = document.getElementById("player-hand")
const currentTrickDiv = document.getElementById("current-trick")
const team1ScoreCurrentRoundSpan = document.getElementById("team1-score-current-round")
const team2ScoreCurrentRoundSpan = document.getElementById("team2-score-current-round")
const team1ScoreTotalSpan = document.getElementById("team1-score-total")
const team2ScoreTotalSpan = document.getElementById("team2-score-total")
const myPlayerNameSpan = document.getElementById("my-player-name")
const teamToggle = document.getElementById("team-toggle")
const pointsGoal = document.getElementById("points-goal-input")
const pointsGoalDisplay = document.getElementById("points-goal")
const declarationArea = document.getElementById("declaration-button-area")
const declarationsSection = document.getElementById("declarations-section")
const declarationInfo = document.getElementById("declaration-info")
const rulesSection = document.getElementById("rules")
const rulesButton = document.getElementById("rules-button")

rulesSection.innerHTML = rules_text // Set the rules text
rulesCloseButton = document.createElement("button")
rulesCloseButtonContainer = document.createElement("div")
rulesCloseButton.textContent = "Close"
rulesCloseButton.classList.add("rules-close-button")
rulesCloseButtonContainer.classList.add("rules-close-button-container")
rulesCloseButtonContainer.appendChild(rulesCloseButton)
rulesSection.appendChild(rulesCloseButtonContainer) // Append the close button to the rules section

rulesCloseButton.addEventListener("click", () => {
    rulesSection.style.display = "none" // Hide the rules section
    rulesButton.textContent = "Show Rules"
})

declarationsSection.addEventListener("click", (event) => {
    declarationsSection.style.display = "none" // Hide the declaration area
})

rulesButton.addEventListener("click", () => {
    if (rulesSection.style.display === "none" || rulesSection.style.display === "") {
        rulesSection.style.display = "block"
        rulesButton.textContent = "Hide Rules"
    }
    else {
        rulesSection.style.display = "none"
        rulesButton.textContent = "Show Rules"
    }
})

// New UI Elements
const initialSection = document.getElementById("initial-section")
const waitingSection = document.getElementById("waiting-section")
const gameContainer = document.getElementById("game-container")
const playerNameInput = document.getElementById("player-name-input")
const createGameButton = document.getElementById("create-game-button")
const createdGameCodeDisplay = document.getElementById("created-game-code-display")
const joinGameCodeInput = document.getElementById("join-game-code-input")
const joinGameButton = document.getElementById("join-game-button")
const gameCodeDisplay = document.getElementById("game-code-display")
const waitingStatus = document.getElementById("waiting-status")
const lobbyPlayersDiv = document.getElementById("lobby-players")

const suitOrder = { Bastoni: 1, Kope: 2, Denari: 3, Spade: 4 }

const playerPositions = {}

// --- Initialization ---

document.addEventListener("DOMContentLoaded", () => {
    // Connect WebSocket on load, but don't join immediately
    connectWebSocket()

    // Add event listeners for create/join buttons
    // Ensure buttons exist before adding listeners
    if (createGameButton) {
        createGameButton.addEventListener("click", createGame)
    }
    if (joinGameButton) {
        joinGameButton.addEventListener("click", joinGame)
    }

    // Initial UI state
    showSection("initial-section")

    addToggleButtonHandlers() // Add toggle button handlers

    const selectedToggle = document.querySelector(".selected")
    if (selectedToggle) {
        if (selectedToggle.id === "red") {
            selectedToggle.classList.add("red-team-selected")
        }
        if (selectedToggle.id === "blue") {
            selectedToggle.classList.add("blue-team-selected")
        }
    }
})

// --- WebSocket Functions ---

function connectWebSocket() {
    // Determine WebSocket protocol (ws or wss)
    const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:"
    const wsUrl = `${wsProtocol}//${window.location.host}/ws`

    ws = new WebSocket(wsUrl)

    ws.onopen = () => {
        console.log("WebSocket connection established")
        statusMessage.textContent = "Connected. Create or join a game."
    }

    ws.onmessage = (event) => {
        try {
            const message = JSON.parse(event.data)
            handleMessage(message)
        } catch (error) {
            console.error("Failed to parse message or handle:", error)
            statusMessage.textContent = "Error processing message from server."
        }
    }

    ws.onerror = (error) => {
        console.error("WebSocket error:", error)
        statusMessage.textContent = "WebSocket connection error."
        showSection("initial-section")
    }

    ws.onclose = () => {
        console.log("WebSocket connection closed")
        statusMessage.textContent = "Disconnected. Please refresh to reconnect."
        showSection("initial-section")
    }
}

function sendMessage(type, payload) {
    if (ws && ws.readyState === WebSocket.OPEN) {
        const message = JSON.stringify({ type, payload })
        if (type !== "ping") {
            console.log(`Sending message (type: ${type}), Payload: `, payload)
        }
        ws.send(message)
    } else {
        console.error("WebSocket is not connected.")
        statusMessage.textContent = "Not connected to server."
        showSection("initial-section")
    }
}

// --- Game Creation and Joining ---

function createGame() {
    const name = playerNameInput.value.trim()
    const pointsGoalValue = pointsGoal.value.trim()
    const selectedTeam = document.querySelector(".selected")
    if (!name) {
        alert("Please enter your name.")
        return
    }
    if (!pointsGoalValue) {
        alert("Please enter the points goal.")
        return
    }
    if (!selectedTeam) {
        alert("Please select a team.")
        return
    }
    const team = selectedTeam.id === "red" ? 1 : 2 // Map team ID to team number
    myPlayerName = name
    sendMessage("create_game", { name, desired_team: team, points_goal: parseInt(pointsGoalValue) }) // Send team ID to server
    waitingStatus.textContent = "Creating game..."
    // Clear join code input if user clicks create after typing in join
    if (joinGameCodeInput) joinGameCodeInput.value = ""
}

function joinGame() {
    const name = playerNameInput.value.trim()
    const gameCode = joinGameCodeInput.value.trim().toUpperCase()
    const desired_team = document.querySelector(".selected") // Get the selected team
    if (!name) {
        alert("Please enter your name.")
        return
    }
    if (!gameCode) {
        alert("Please enter the game code to join.")
        return
    }
    if (!desired_team) {
        alert("Please select a team.")
        return
    }
    const team = desired_team.id === "red" ? 1 : 2 // Map team ID to team number
    myPlayerName = name
    sendMessage("join_game", { name, game_code: gameCode, desired_team: team }) // Send team ID to server
    showSection("waiting-section") // Switch to waiting section on attempting join
    waitingStatus.textContent = "Joining game..."
    gameCodeDisplay.textContent = gameCode
    // Clear create code display if user clicks join after creating
    if (createdGameCodeDisplay) createdGameCodeDisplay.value = ""
}

// --- Message Handling ---

function handleMessage(message) {
    // Log all messages except pong for debugging
    // Not logging pong to reduce noise in console
    if (message.type !== "pong") {
        console.log("Handling message:", message)
    }
    switch (message.type) {
        case "game_created":
            handleGameCreated(message.payload)
            break
        case "lobby_update":
            handleLobbyUpdate(message.payload)
            break
        case "join_error":
            handleJoinError(message.payload)
            break
        case "game_start":
            handleGameStart(message.payload)
            break
        case "deal_hand":
            handleDealHand(message.payload)
            break
        case "your_turn":
            handleYourTurn()
            break
        case "game_state_update":
            handleGameState(message.payload)
            break
        case "you_played":
            handlePlayerPlayedCard(message.payload)
            break
        case "trick_end":
            handleTrickEnd(message.payload)
            break
        case "round_end":
            handleRoundEnd(message.payload)
            break
        case "game_over":
            handleGameOver(message.payload)
            break
        case "declaration_confirmation":
            handleDeclarationConfirmation(message.payload)
            break
        case "error":
            handleGenericError(message.payload)
            break
        case "pong":
            break
        default:
            console.warn("Received unhandled message type:", message.type)
    }
}

function handleGameCreated(payload) {
    if (createdGameCodeDisplay) {
        createdGameCodeDisplay.value = payload.game_code // Display the created game code
    }
    gameCodeDisplay.textContent = payload.game_code // Also update lobby display
    showSection("waiting-section") // Now switch to waiting section
    waitingStatus.textContent = "Waiting for other players to join..."
}

function handleLobbyUpdate(payload) {
    lobbyPlayersDiv.innerHTML = "" // Clear previous list
    payload.players.forEach((player) => {
        const playerElement = document.createElement("div")
        playerElement.textContent = player.name + (player.name === myPlayerName ? " (You)" : "")
        lobbyPlayersDiv.appendChild(playerElement)
    })
    waitingStatus.textContent = `Waiting for players (${payload.players.length}/4)...`
}

function handleJoinError(payload) {
    console.error("Failed to join game:", payload.message)
    alert(`Join Error: ${payload.message}`)
    showSection("initial-section") // Go back to initial screen
}

function handleGenericError(payload) {
    console.error("Server error:", payload.message)
    statusMessage.textContent = `Error: ${payload.message}`
}

function handleGameStart(payload) {
    myPlayerId = findPlayerIdByName(payload.players, myPlayerName) // Find our player ID based on name
    playerPositions[myPlayerId] = "player-bottom" // Assign position for our hand
    if (!myPlayerId) {
        console.error("Could not find own player ID in game_start payload!")
        // Handle this error appropriately - maybe disconnect?
    }
    teamsInfo = payload.teams // Store teams info for later use
    showSection("game-container")
    myPlayerNameSpan.textContent = myPlayerName
    const teamIndex = getTeamIndexByPlayerId(myPlayerId)
    if (teamIndex !== -1) {
        myPlayerNameSpan.classList.add(`team${teamIndex}`)
    } else {
        console.error("Could not find my team index!")
    }
    setupOpponentNames(payload.players, payload.teams)
    pointsGoalDisplay.textContent = `Points Goal: ${payload.points_goal}`
    canDeclare = true
    resetScores()
}

function handleDealHand(payload) {
    statusMessage.textContent = "Cards dealt. Waiting for first turn."
    handCards = payload.hand // Store hand cards for later use
    renderHand(payload.hand)
}

function handleYourTurn() {
    statusMessage.textContent = "Your turn!"
    highlightPlayableCards()
    renderDeclarations()
    removeDeclarationInfo()
}

function handleGameState(payload) {
    const currentPlayer = teamsInfo
        .find((t) => t.players.some((p) => p.id === payload.current_player_id))
        .players.find((p) => p.id === payload.current_player_id)
    document.querySelectorAll(".current-player").forEach((el) => el.classList.remove("current-player")) // Remove previous highlights
    document.getElementById(playerPositions[payload.current_player_id]).querySelector(".player-name").classList.add("current-player") // Highlight current player
    const playerName = currentPlayer.name || payload.current_player_id // Fallback to ID if name not found
    if (currentPlayer.id !== myPlayerId && !afterTrick) {
        statusMessage.textContent = `${playerName}'s turn` // Update based on actual name later
    } else if (afterTrick) {
        afterTrick = false // Reset after trick flag
    }
    renderTrick(payload.cards_on_table)
    trickCards = payload.cards_on_table // Store cards in the current trick
}

function handlePlayerPlayedCard(payload) {
    if (payload.player_id === myPlayerId) {
        canDeclare = false // Disable declaration after playing a card
        removeCardFromHand(payload.card)
        removeHighlightedCards() // Remove highlight from all cards
        renderDeclarations() // Update declarations section
    }
}

function handleTrickEnd(payload) {
    afterTrick = true // Set flag to indicate trick has ended
    playDisabled = true // Disable play card action until trick is cleared
    updateScoresAfterTrick(payload.winner_id, payload.points) // Update scores based on trick results
    const playerNameMaybe = teamsInfo
        .find((t) => t.players.some((p) => p.id === payload.winner_id))
        .players.find((p) => p.id === payload.winner_id).name
    const playerName = playerNameMaybe || payload.winner_id // Fallback to ID if name not found
    if (payload.winner_id === myPlayerId) {
        statusMessage.textContent = "You won the trick!"
    } else {
        statusMessage.textContent = `${playerName} won the trick!`
    }
    // Clear the trick area after a short delay
    setTimeout(() => {
        clearTrickDisplay()
        if (gameOver) {
            return
        }
        if (payload.winner_id !== myPlayerId) {
            statusMessage.textContent = `Waiting for ${playerName} to lead next trick.`
        }
        playDisabled = false // Re-enable play card action
        if (roundOver) {
            statusMessage.textContent = "Round over. Waiting for next round."
            team1ScoreTotalSpan.textContent = `${roundOverPayload.team1_total_score}`
            team2ScoreTotalSpan.textContent = `${roundOverPayload.team2_total_score}`
            // Reset round over flag
            roundOver = false
            roundOverPayload = null // Clear the payload
            resetScores() // Reset scores for the next round
        }
    }, 5000) // 2-second delay
    // Make a glow effect on the winning card after 500 ms
    setTimeout(() => {
        const winningCard = currentTrickDiv.querySelector(`[data-card-id="${payload.winner.card.Suit}-${payload.winner.card.Rank}"]`)
        if (winningCard) {
            winningCard.classList.add("glow")
        }
    }, 500)
}

function handleRoundEnd(payload) {
    roundOver = true // Set flag to indicate round has ended
    roundOverPayload = payload // Store the payload for round over
}

function handleGameOver(payload) {
    // TODO: show team name instead of ID
    gameOver = true // Set flag to indicate game is over
    statusMessage.textContent = `Game Over! Winning Team: ${payload.winning_team_id}. Final Score: T1 ${payload.final_score_t1} - T2 ${payload.final_score_t2}`
    playerHandDiv.innerHTML = "<p>Game Over</p>"
    currentTrickDiv.innerHTML = ""
    setTimeout(() => {
        showSection("initial-section")
        myPlayerId = null // Reset player ID
        myPlayerName = null // Reset player name
        teamsInfo = null // Reset teams info
        trickCards = [] // Reset trick cards
        handCards = [] // Reset hand cards
        playDisabled = false // Reset play disabled flag
        afterTrick = false // Reset after trick flag
        roundOver = false // Reset round over flag
        roundOverPayload = null // Reset round over payload
        gameOver = false // Reset game over flag
    }, 10000)
}

// --- UI Rendering Functions ---

function addToggleButtonHandlers() {
    teamToggle.addEventListener("click", () => {
        const selected = document.querySelector(".selected")
        const notSelected = document.querySelector(".not-selected")
        if (selected && notSelected) {
            selected.classList.remove("selected")
            selected.classList.add("not-selected")
            notSelected.classList.remove("not-selected")
            notSelected.classList.add("selected")
            if (notSelected.id === "red") {
                notSelected.classList.add("red-team-selected")
                selected.classList.remove("blue-team-selected")
            } else if (notSelected.id === "blue") {
                notSelected.classList.add("blue-team-selected")
                selected.classList.remove("red-team-selected")
            }
        }
    })
}

function showSection(sectionId) {
    initialSection.classList.add("hidden")
    waitingSection.classList.add("hidden")
    gameContainer.classList.add("hidden")

    const sectionToShow = document.getElementById(sectionId)
    if (sectionToShow) {
        sectionToShow.classList.remove("hidden")
    } else {
        console.error(`Section with ID ${sectionId} not found.`)
    }
}

function renderHand(hand) {
    playerHandDiv.innerHTML = "" // Clear previous hand
    hand.sort(compareCards) // Sort hand by suit and rank
    hand.forEach((card) => {
        const cardElement = createCardElement(card)
        cardElement.addEventListener("click", () => playCard(card))
        playerHandDiv.appendChild(cardElement)
    })
}

function renderTrick(cards) {
    // Check if trick is empty. If so return
    if (cards.length === 0) {
        return
    }
    currentTrickDiv.innerHTML = "" // Clear previous trick
    cards.forEach((card) => {
        const cardElement = createCardElement(card, true)
        currentTrickDiv.appendChild(cardElement)
    })
}

function clearTrickDisplay() {
    currentTrickDiv.innerHTML = ""
    // for (let i = 0; i < 4; i++) {
    //     const placeholder = document.createElement('div');
    //     placeholder.classList.add('card-placeholder', 'trick-card');
    //     currentTrickDiv.appendChild(placeholder);
    // }
}

function renderDeclarations() {
    if (canDeclare) {
        const declareButton = document.createElement("button")
        const napolaDeclarations = document.createElement("div")
        const napolaDeclarationsTitle = document.createElement("h3")
        const threeFourOfAKindDeclarations = document.createElement("div")
        const threeFourOfAKindDeclarationsTitle = document.createElement("h3")
        declareButton.textContent = "Declare"
        declareButton.classList.add("declare-button")
        napolaDeclarations.classList.add("declarations-section-inner")
        threeFourOfAKindDeclarations.classList.add("declarations-section-inner")
        napolaDeclarationsTitle.textContent = "Napola Declarations"
        threeFourOfAKindDeclarationsTitle.textContent = "Three or Four of a Kind Declarations"
        declarations.forEach((declaration) => {
            const declarationElement = document.createElement("img")
            const type = declaration.type
            let suit = ""
            let rank = ""
            declarationElement.classList.add("card")
            declarationElement.style.cursor = "pointer"
            declarationElement.src = declaration.image
            if (declaration.type === "napola") {
                suit = declaration.suit
                napolaDeclarations.appendChild(declarationElement)
            } else if (declaration.type === "three_or_four_of_kind") {
                rank = declaration.rank
                threeFourOfAKindDeclarations.appendChild(declarationElement)
            }

            declarationElement.addEventListener("click", () => {
                sendMessage("declare", { declaration_type: type, suit: suit, rank: rank })
            })
        })
        declarationsSection.appendChild(napolaDeclarationsTitle)
        declarationsSection.appendChild(napolaDeclarations)
        declarationsSection.appendChild(threeFourOfAKindDeclarationsTitle)
        declarationsSection.appendChild(threeFourOfAKindDeclarations)

        declareButton.addEventListener("click", () => {
            // declarationsSection has display: none; so we need to show it
            declarationsSection.style.display = "block"
        })
        declarationArea.appendChild(declareButton)
    } else {
        const declareButton = document.querySelector(".declare-button")
        if (declareButton) {
            declareButton.remove() // Remove the button if it exists
        }
        declarationsSection.style.display = "none" // Hide the declaration area
    }
}

function removeCardFromHand(cardToRemove) {
    const cardId = `${cardToRemove.Suit}-${cardToRemove.Rank}`
    const cardElement = playerHandDiv.querySelector(`[data-card-id="${cardId}"]`)
    if (cardElement) {
        cardElement.remove()
        handCards = handCards.filter((card) => card.Suit !== cardToRemove.Suit || card.Rank !== cardToRemove.Rank) // Update hand cards array
    } else {
        console.warn(`Could not find card ${cardId} in hand to remove.`)
    }
}

function createCardElement(card, isTrickCard = false) {
    const cardElement = document.createElement("img")
    cardElement.classList.add("card")
    if (isTrickCard) {
        cardElement.classList.add("trick-card")
    }
    // Use suit and rank to set the src attribute
    const imageName = `${card.Suit}_${card.Rank}.png`
    cardElement.src = `images/cards/${imageName}`
    cardElement.dataset.suit = card.Suit
    cardElement.dataset.rank = card.Rank
    cardElement.dataset.cardId = `${card.Suit}-${card.Rank}`
    cardElement.alt = `${card.Rank} of ${card.Suit}`
    cardElement.title = `${card.Rank} of ${card.Suit}`
    return cardElement
}

function highlightPlayableCards() {
    const validMoves = trickCards.length === 0 ? handCards : handCards.filter((card) => trickCards[0].Suit === card.Suit)

    if (validMoves.length === 0) {
        validMoves.push(...handCards) // If no valid moves, all cards are playable
    }

    removeHighlightedCards() // Clear previous highlights

    if (validMoves && validMoves.length > 0) {
        validMoves.forEach((card) => {
            const cardId = `${card.Suit}-${card.Rank}`
            const cardElement = playerHandDiv.querySelector(`[data-card-id="${cardId}"]`)
            if (cardElement) {
                cardElement.classList.add("playable")
            }
        })
    }
}

function removeHighlightedCards() {
    // Remove highlight from all cards
    playerHandDiv.querySelectorAll(".card").forEach((cardEl) => {
        cardEl.classList.remove("playable")
    })
}

function resetScores() {
    // Reset scores for both teams to 0
    teamsInfo.forEach((team) => {
        if (team.team_number === 1) {
            team1ScoreCurrentRoundSpan.textContent = "0"
        } else if (team.team_number === 2) {
            team2ScoreCurrentRoundSpan.textContent = "0"
        }
    })

    teamsInfo.forEach((team) => {
        team.score = 0 // Reset score for each team
    })
}

function updateScoresAfterTrick(playerId, points) {
    // Update scores based on trick results
    teamsInfo.forEach((team) => {
        if (team.players.some((p) => p.id === playerId)) {
            team.score += points // Add points to the winning team
            if (team.team_number === 1) {
                const team1Score = team.score % 3 === 0 ? `${team.score / 3}` : `${team.score} / 3`
                team1ScoreCurrentRoundSpan.textContent = `${team1Score}`
            } else if (team.team_number === 2) {
                const team2Score = team.score % 3 === 0 ? `${team.score / 3}` : `${team.score} / 3`
                team2ScoreCurrentRoundSpan.textContent = `${team2Score}`
            }
        }
    })
}

function handleDeclarationConfirmation(payload) {
    updateScoresAfterDeclarationConfirmation(payload)
    // find player name ('You' if it's me)
    const playerName = payload.player_id === myPlayerId ? "You" : teamsInfo
        .find((t) => t.players.some((p) => p.id === payload.player_id))
        .players.find((p) => p.id === payload.player_id).name

    let message = `${playerName} declared `
    if (payload.declaration.declaration_type === "napola") {
        message += `Napola ${payload.declaration.suit}`
    } else if (payload.declaration.declaration_type === "three_or_four_of_kind") {
        if (payload.points === 3 * 3) {
            message += `Three of a Kind ${payload.declaration.rank} without ${payload.without_suit}`
        } else if (payload.points === 4 * 3) {
            message += `Four of a Kind ${payload.declaration.rank}`
        }
    }

    if (payload.points > 0) {
        declarationInfo.textContent = message
        declarationInfo.style.display = "block"
        declarationInfo.classList.add("declaration-info")
        setTimeout(() => {
            removeDeclarationInfo()
        }, 5000) // Clear message after 5 seconds
    }
}

function removeDeclarationInfo() {
    declarationInfo.textContent = ""
    declarationInfo.style.display = "none"
    declarationInfo.classList.remove("declaration-info")
}

function updateScoresAfterDeclarationConfirmation(payload) {
    // Update scores based on declaration confirmation
    teamsInfo.forEach((team) => {
        if (team.id === payload.team_id) {
            team.score += payload.points // Add points to the winning team
            if (team.team_number === 1) {
                const team1Score = team.score % 3 === 0 ? `${team.score / 3}` : `${team.score} / 3`
                team1ScoreCurrentRoundSpan.textContent = `${team1Score}`
            } else if (team.team_number === 2) {
                const team2Score = team.score % 3 === 0 ? `${team.score / 3}` : `${team.score} / 3`
                team2ScoreCurrentRoundSpan.textContent = `${team2Score}`
            }
        }
    })
}

function setupOpponentNames(players, teams) {
    // Find our team and opponents
    let myTeamNumber = -1
    let myTeamId = null
    teams.forEach((team) => {
        if (team.players.some((p) => p.id === myPlayerId)) {
            myTeamNumber = team.team_number
            myTeamId = team.id
        }
    })

    if (myTeamNumber === -1) {
        console.error("Could not determine player's team!")
        return
    }

    const partner = players.find(
        (p) => p.id !== myPlayerId && teams.some((t) => t.id === myTeamId && t.players.some((tp) => tp.id === p.id))
    )
    const opponents = players.filter(
        (p) => p.id !== myPlayerId && !teams.some((t) => t.id === myTeamId && t.players.some((tp) => tp.id === p.id))
    )

    const opponentLeft = opponents.find((p) => p.position === (partner.position + 1) % 4)
    const opponentRight = opponents.find((p) => p.position === (partner.position === 0 ? 3 : partner.position - 1))

    if (partner) {
        const partnerSpan = document.querySelector("#opponent-top .player-name")
        if (partnerSpan) {
            partnerSpan.textContent = partner.name
            playerPositions[partner.id] = "opponent-top" // Assign position
            const partnerTeamIndex = getTeamIndexByPlayerId(partner.id)
            if (partnerTeamIndex !== -1) {
                partnerSpan.classList.add(`team${partnerTeamIndex}`) // Add team class for styling
            } else {
                console.error("Could not find partner's team index!")
            }
        } else {
            console.error("Partner name span not found!")
        }
    }
    if (opponents.length === 2) {
        const opponentLeftSpan = document.querySelector("#opponent-left .player-name")
        const opponentRightSpan = document.querySelector("#opponent-right .player-name")
        if (opponentLeftSpan && opponentRightSpan) {
            opponentLeftSpan.textContent = opponentLeft.name
            opponentRightSpan.textContent = opponentRight.name
            playerPositions[opponentLeft.id] = "opponent-left" // Assign position
            playerPositions[opponentRight.id] = "opponent-right" // Assign position
            // searching for only one opponent team index, assuming both are on the same team
            const opponentTeamIndex = getTeamIndexByPlayerId(opponents[0].id)
            if (opponentTeamIndex !== -1) {
                opponentLeftSpan.classList.add(`team${opponentTeamIndex}`) // Add team class for styling
            }
            if (opponentTeamIndex !== -1) {
                opponentRightSpan.classList.add(`team${opponentTeamIndex}`) // Add team class for styling
            }
        } else {
            console.error("Opponent name spans not found!")
        }
    }
}

// --- Player Actions ---

function playCard(card) {
    if (playDisabled) {
        return
    }
    sendMessage("play_card", { suit: card.Suit, rank: card.Rank })
}

// --- Utility Functions ---

// Find player ID based on name (needed because server might not send our ID initially)
function findPlayerIdByName(players, name) {
    const player = players.find((p) => p.name === name)
    return player ? player.id : null
}

// Compare cards by suit and rank (or order)
function compareCards(cardA, cardB) {
    const suitsDiff = suitOrder[cardA.Suit] - suitOrder[cardB.Suit]
    if (suitsDiff !== 0) {
        return suitsDiff // Compare by suit first
    }

    return cardA.Order - cardB.Order // Then by rank (or order)
}

function getTeamIndexByPlayerId(playerId) {
    for (let i = 0; i < teamsInfo.length; i++) {
        if (teamsInfo[i].players.some((p) => p.id === playerId)) {
            return i + 1 // Return the index of the team containing the player
        }
    }
    return -1 // Not found
}

// Keepalive using ping/pong
setInterval(() => {
    if (ws && ws.readyState === WebSocket.OPEN) {
        sendMessage("ping", {})
    }
}, 30000) // Send ping every 30 seconds
