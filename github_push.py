import getpass
import os
import subprocess
import sys
from pathlib import Path

PAT_FILE = Path(".pat")
USERNAME_FILE = Path(".username")

def get_saved_token():
    if PAT_FILE.exists():
        token = PAT_FILE.read_text().strip()
        if token:
            print(f"Loaded token from {PAT_FILE.name}")
            return token
    return None

def save_token(token):
    try:
        PAT_FILE.write_text(token)
        print(f"Token saved to {PAT_FILE.name}")
    except Exception as e:
        print(f"Warning: Could not save token to file: {e}")

def get_saved_username():
    if USERNAME_FILE.exists():
        name = USERNAME_FILE.read_text().strip()
        if name:
            print(f"Loaded username from {USERNAME_FILE.name}")
            return name
    return None

def save_username(name):
    try:
        USERNAME_FILE.write_text(name)
        print(f"Username saved to {USERNAME_FILE.name}")
    except Exception as e:
        print(f"Warning: Could not save username to file: {e}")

def get_git_remote_url():
    try:
        return subprocess.getoutput("git remote get-url origin").strip()
    except Exception:
        return ""

def commit_changes(message=None):
    # Check status
    try:
        status = subprocess.getoutput("git status --porcelain").strip()
    except Exception:
         print("Error getting git status.")
         return

    if not status:
        print("No changes to commit.")
        return

    print("\n--- Committing Changes ---")
    print("Files changed:")
    print(status)
    
    if message:
        print(f"Auto-committing with message: {message}")
        subprocess.run(["git", "add", "."], check=True)
        subprocess.run(["git", "commit", "-m", message], check=True)
    else:
        confirm = input("Do you want to commit these changes? (Y/n): ").strip().lower()
        if confirm != 'n':
            subprocess.run(["git", "add", "."], check=True)
            message = input("Enter commit message: ").strip()
            if not message:
                message = "Update"
            subprocess.run(["git", "commit", "-m", message], check=True)

def push_changes(username, pat, repo_name):
    # Construct the authed URL
    # Using the token directly in the URL
    remote_url = f"https://{username}:{pat}@github.com/{username}/{repo_name}.git"

    # Get current branch
    current_branch = subprocess.getoutput("git branch --show-current").strip()
    if not current_branch:
         print("⚠️ Could not detect current branch. Defaulting to 'main'.")
         current_branch = "main"

    print(f"Pushing branch '{current_branch}' to remote '{username}/{repo_name}'...")
    
    # Run git push - we expect errors to be caught by the caller
    process = subprocess.run(["git", "push", remote_url, current_branch], capture_output=True, text=True)
    
    if process.returncode != 0:
        print(process.stderr)
        if "403" in process.stderr or "permission denied" in process.stderr.lower():
             raise PermissionError("Access denied (403). check token permissions.")
        raise subprocess.CalledProcessError(process.returncode, process.args, output=process.stdout, stderr=process.stderr)
    
    print(process.stdout)

def main():
    print("--- GitHub Push Helper ---")
    print("Use this script to configure your git user/email and push with a Personal Access Token (PAT).")
    print("Note: The token and username will be saved locally (and ignored by git).")
    print("--------------------------")

    # 1. Get Username
    default_user = 'chanckjoseph'
    saved_username = get_saved_username()
    
    if saved_username:
        username = saved_username
    else:
        username_input = input(f"Enter GitHub Username [{default_user}]: ").strip()
        username = username_input if username_input else default_user
        save_username(username)

    # 2. Get PAT (Try saved first)
    pat = get_saved_token()
    
    if not pat:
        print("No saved token found.")
        pat = getpass.getpass("Enter Personal Access Token (PAT): ").strip()
        if not pat:
            print("❌ PAT is required to push.")
            sys.exit(1)
        save_token(pat) # Save immediately on regular input

    # 3. Configure Git User/Email (Only if not set properly)
    print("\n--- Git Configuration ---")
    try:
        current_name = subprocess.getoutput("git config user.name").strip()
        print(f"Current git user.name: {current_name if current_name else '(not set)'}")
        
        # Only ask if name looks empty, otherwise skip to keep it fast
        if not current_name:
             new_name = input(f"Enter user.name [{username}]: ").strip() or username
             subprocess.run(["git", "config", "user.name", new_name], check=False)
             
        current_email = subprocess.getoutput("git config user.email").strip()
        if not current_email:
             new_email = input(f"Enter user.email: ").strip()
             if new_email:
                subprocess.run(["git", "config", "user.email", new_email], check=False)
    except Exception:
        pass

    import sys
    commit_msg = None
    if len(sys.argv) > 1:
        commit_msg = " ".join(sys.argv[1:])
    
    commit_changes(commit_msg)
    
    # 4. Construct Remote URL
    default_repo = "personnel-productivity"
    try:
        detect_url = get_git_remote_url()
        # Expecting something like https://github.com/user/repo.git
        if "github.com" in detect_url:
            parts = detect_url.replace(".git", "").split("/")
            if len(parts) >= 1:
                repo_candidate = parts[-1]
                if repo_candidate:
                    default_repo = repo_candidate
    except Exception:
        pass

    repo_name = default_repo
    
    print("\n--- Pushing to GitHub ---")
    
    try:
        push_changes(username, pat, repo_name)
        print("✅ Push successful!")
    except (subprocess.CalledProcessError, PermissionError) as e:
        print(f"\n❌ Push failed: {e}")
        print("The token might be invalid or expired.")
        retry = input("Do you want to enter a new token and retry? (Y/n): ").strip().lower()
        if retry != 'n':
            pat = getpass.getpass("Enter NEW Personal Access Token (PAT): ").strip()
            if pat:
                save_token(pat)
                try:
                    # Retry with new token
                    push_changes(username, pat, repo_name)
                    print("✅ Push successful with new token!")
                except Exception as e2:
                    print(f"❌ Push failed again: {e2}")

if __name__ == "__main__":
    main()
