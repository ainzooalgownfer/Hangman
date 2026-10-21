#!/usr/bin/env bash
# Simulates many concurrent Hangman players against a running server.
#
# Usage: scripts/simulate.sh [players] [host] [port]
#   players defaults to 20, host to localhost, port to 8080.
#
# Each player opens its own TCP connection (one goroutine on the server),
# plays a full game, then quits. The script exits non-zero if any player
# hit an error or didn't finish, or if the server died during the run.

N=${1:-20}
HOST=${2:-localhost}
PORT=${3:-8080}

# Most common letters first, so bots finish in a reasonable number of guesses.
LETTERS="e a r o i t n s l c u d p m h g b f y w k v x z j q"

# send writes one JSON line on fd 3 and reads the one-line response into $resp.
send() {
	printf '%s\n' "$1" >&3
	read -r -t 5 resp <&3 || resp="TIMEOUT"
}

# play_one plays a full game and prints exactly one result line.
play_one() {
	local n=$1 id letter
	exec 3<>"/dev/tcp/$HOST/$PORT" 2>/dev/null || { echo "player $n: CONNECT_FAILED"; return; }

	send "{\"action\":\"NEW_GAME\",\"player\":\"bot$n\"}"
	id=$(sed -nE 's/.*"session_id":"([^"]+)".*/\1/p' <<<"$resp")
	if [[ -z $id ]]; then
		echo "player $n: ERROR no session id: $resp"
		exec 3>&-
		return
	fi

	for letter in $LETTERS; do
		send "{\"action\":\"LETTER\",\"session_id\":\"$id\",\"letter\":\"$letter\"}"
		case $resp in
			*'"status":"WON"'*)  echo "player $n: WON $id";  break ;;
			*'"status":"LOST"'*) echo "player $n: LOST $id"; break ;;
			*'"type":"ERROR"'* | TIMEOUT)
				echo "player $n: ERROR $id $resp"
				exec 3>&-
				return
				;;
		esac
	done

	# Clean up so finished games don't stay in the server's memory.
	send "{\"action\":\"QUIT\",\"session_id\":\"$id\"}"
	exec 3>&-
}

server_alive() {
	(exec 3<>"/dev/tcp/$HOST/$PORT") 2>/dev/null
}

if ! server_alive; then
	echo "No server on $HOST:$PORT. Start it first: go run ./cmd/server" >&2
	exit 1
fi

out=$(mktemp)
trap 'rm -f "$out"' EXIT

echo "Starting $N players against $HOST:$PORT..."
start=${EPOCHREALTIME/./}

for i in $(seq 1 "$N"); do
	play_one "$i" &
done >"$out"
wait

end=${EPOCHREALTIME/./}
elapsed_ms=$(((end - start) / 1000))

won=$(grep -c ': WON' "$out")
lost=$(grep -c ': LOST' "$out")
errors=$(grep -c -E ': (ERROR|CONNECT_FAILED)' "$out")
finished=$((won + lost))
missing=$((N - $(wc -l <"$out")))

echo "-----------------------------"
echo "players:  $N"
echo "won:      $won"
echo "lost:     $lost"
echo "errors:   $errors"
echo "no reply: $missing"
echo "time:     ${elapsed_ms} ms"
echo "-----------------------------"

if ((errors > 0)); then
	echo "Errors:"
	grep -E ': (ERROR|CONNECT_FAILED)' "$out" | head -10
fi

if ! server_alive; then
	echo "FAIL: server is no longer reachable" >&2
	exit 1
fi

if ((finished != N)); then
	echo "FAIL: $((N - finished)) of $N players did not finish" >&2
	exit 1
fi

echo "OK: all $N games finished, server still up"
