import getpass
import os
import subprocess
import sys

def main():
    print("--- GitHub Push Helper ---")
    print("Use this script to configure your git user/email and push with a Personal Access Token (PAT).")
    print("Note: The token is NOT saved to disk. It is only used for this session.")
    print("--------------------------")

    # 1. Get Username
    default_user = 'chanckjoseph'
    username_input = input(f"Enter GitHub Username [{default_user}]: ").strip()
    username = username_input if username_input else default_user

    # 2. Get PAT
    pat = getpass.getpass("Enter Personal Access Token (PAT): ").strip()
    if not pat:
        print("❌ PAT is required to push.")
        sys.exit(1)

    # 3. Configure Git User/Email
    print("\n--- Git Configuration ---")
    
    try:
        current_name = subprocess.getoutput("git config user.name").strip()
        current_email = subprocess.getoutput("git config user.email").strip()
        print(f"Current git user.name: {current_name}")
        print(f"Current git user.email: {current_email}")
    except Exception:
        pass
    
    confirm = input("Update git config user.name and user.email? (y/N): ").strip().lower()
    
    if confirm == 'y':
        new_name_input = input(f"Enter user.name [{username}]: ").strip()
        new_name = new_name_input if new_name_input else username
        
        new_email = input(f"Enter user.email: ").strip()
        
        subprocess.run(["git", "config", "user.name", new_name], check=False)
        if new_email:
            subprocess.run(["git", "config", "user.email", new_email], check=False)

    # 4. Construct Remote URL
    # Format: https://USERNAME:TOKEN@github.com/USERNAME/REPO.git
    
    # Try to detect repo name
    default_repo = "personnel-productivity"
    try:
        detect_url = subprocess.getoutput("git remote get-url origin").strip()
        # Expecting something like https://github.com/user/repo.git
        if "github.com" in detect_url:
            parts = detect_url.replace(".git", "").split("/")
            if len(parts) >= 1:
                default_repo = parts[-1]
    except Exception:
        pass

    repo_input = input(f"Enter repository name [{default_repo}]: ").strip()
    repo_name = repo_input if repo_input else default_repo

    # Construct the authed URL
    # Using the token directly in the URL
    remote_url = f"https://{username}:{pat}@github.com/{username}/{repo_name}.git"
    
    print("\n--- Pushing to GitHub ---")
    
    # 5. Run git push
    try:
        current_branch = subprocess.getoutput("git branch --show-current").strip()
        if not current_branch:
             print("⚠️ Could not detect current branch. Defaulting to 'main'.")
             current_branch = "main"

        print(f"Pushing branch '{current_branch}' to remote '{username}/{repo_name}'...")
        
        # We push to the URL directly
        # Hide the sensitive URL from print output but pass it directly
        subprocess.run(["git", "push", remote_url, current_branch], check=True)
        print("✅ Push successful!")
    except subprocess.CalledProcessError:
        print("❌ Push failed. Check your token, permissions, and internet connection.")
    except FileNotFoundError:
        print("❌ Git not found. Please install git.")

if __name__ == "__main__":
    main()
