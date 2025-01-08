import json
import os

# Directory for tournament files
TOURNAMENT_DIR = "tournaments"

def ensure_tournament_dir():
    if not os.path.exists(TOURNAMENT_DIR):
        os.makedirs(TOURNAMENT_DIR)

def load_tournament(tournament_name):
    tournament_file = os.path.join(TOURNAMENT_DIR, tournament_name)
    if os.path.exists(tournament_file):
        with open(tournament_file, 'r') as f:
            return json.load(f), tournament_file
    else:
        print(f"Tournament file '{tournament_file}' not found. Starting a new tournament.")
        return {"all_rounds_no": 5, "rounds": [], "players": []}, tournament_file

def save_tournament(tournament, tournament_file):
    with open(tournament_file, 'w') as f:
        json.dump(tournament, f, indent=4)
    print(f"Tournament saved to {tournament_file}")

def display_tournament(tournament_file):
    with open(tournament_file, 'r', encoding='utf-8') as file:
        tournament = json.load(file)
    print(json.dumps(tournament, indent=4, ensure_ascii=False))

def add_player(tournament):
    identity = input("Enter player name: ")
    start_no = len(tournament['players']) + 1
    tournament['players'].append({"start_no": start_no, "identity": identity, "withdrawals": []})

def enter_results(tournament):
    round_no = int(input("Enter round number: "))
    results = input("Enter results (e.g., 1-0, 0.5-0.5): ")
    tournament['rounds'].append({"round_no": round_no, "results": results})

def generate_pairing(tournament_file):
    wrapper_path = "../src/pairing_engine/wrapper/wrapper.go"
    os.system(f"go run {wrapper_path} {tournament_file}")

def main():
    ensure_tournament_dir()
    tournament_name = input(f"Enter tournament file name (stored in '{TOURNAMENT_DIR}' folder): ")
    tournament_name = tournament_name if tournament_name.endswith('.json') else f"{tournament_name}.json"

    tournament, tournament_file = load_tournament(tournament_name)

    while True:
        print("\n1. View Tournament")
        print("2. Add Player")
        print("3. Enter Results")
        print("4. Generate Pairing")
        print("5. Save and Exit")
        choice = input("Choose an option: ")
        if choice == '1':
            display_tournament(tournament_file)
        elif choice == '2':
            add_player(tournament)
            save_tournament(tournament, tournament_file)
        elif choice == '3':
            enter_results(tournament)
            save_tournament(tournament, tournament_file)
        elif choice == '4':
            generate_pairing(tournament_file)
            tournament, tournament_file = load_tournament(tournament_name)
        elif choice == '5':
            save_tournament(tournament, tournament_file)
            break
        else:
            print("Invalid choice.")

if __name__ == "__main__":
    main()
