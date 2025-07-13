package shell

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// getRepository は現在のディレクトリのGitリポジトリを取得します
func (s *Shell) getRepository() (*git.Repository, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainOpen(wd)
	if err != nil {
		return nil, fmt.Errorf("gitリポジトリが見つかりません: %v", err)
	}
	return repo, nil
}

// gitStatus はGitリポジトリの状態を表示します
func (s *Shell) gitStatus(args []string) error {
	repo, err := s.getRepository()
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("ワークツリーの取得に失敗しました: %v", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("ステータスの取得に失敗しました: %v", err)
	}

	if status.IsClean() {
		fmt.Println("作業ディレクトリはクリーンです")
		return nil
	}

	fmt.Println("変更されたファイル:")
	for file, fileStatus := range status {
		var statusText string
		switch {
		case fileStatus.Staging == git.Untracked:
			statusText = "?? " // 未追跡
		case fileStatus.Staging == git.Added:
			statusText = "A  " // 追加（ステージング済み）
		case fileStatus.Staging == git.Modified:
			statusText = "M  " // 変更（ステージング済み）
		case fileStatus.Worktree == git.Modified:
			statusText = " M " // 変更（未ステージング）
		case fileStatus.Worktree == git.Deleted:
			statusText = " D " // 削除（未ステージング）
		default:
			statusText = "   "
		}
		fmt.Printf("%s %s\n", statusText, file)
	}
	return nil
}

// gitAdd はファイルをステージングエリアに追加します
func (s *Shell) gitAdd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("エラー: 追加するファイルを指定してください")
	}

	repo, err := s.getRepository()
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("ワークツリーの取得に失敗しました: %v", err)
	}

	for _, file := range args {
		if file == "." {
			err = worktree.AddGlob("*")
		} else {
			_, err = worktree.Add(file)
		}
		if err != nil {
			fmt.Printf("エラー: ファイル '%s' を追加できませんでした: %v\n", file, err)
			continue
		}
		fmt.Printf("ファイルを追加しました: %s\n", file)
	}
	return nil
}

// gitCommit は変更をコミットします
func (s *Shell) gitCommit(args []string) error {
	if len(args) < 2 || args[0] != "-m" {
		return fmt.Errorf("エラー: コミットメッセージを指定してください\n使用法: git commit -m \"コミットメッセージ\"")
	}

	repo, err := s.getRepository()
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("ワークツリーの取得に失敗しました: %v", err)
	}

	commit, err := worktree.Commit(args[1], &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Git CLI User",
			Email: "user@git-cli.local",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("コミットに失敗しました: %v", err)
	}

	fmt.Printf("コミットが作成されました: %s\n", commit.String()[:8])
	return nil
}

// gitPush はリモートリポジトリにプッシュします
func (s *Shell) gitPush(args []string) error {
	repo, err := s.getRepository()
	if err != nil {
		return err
	}

	err = repo.Push(&git.PushOptions{})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			fmt.Println("既に最新の状態です")
			return nil
		}
		return fmt.Errorf("プッシュに失敗しました: %v", err)
	}

	fmt.Println("プッシュが完了しました")
	return nil
}

// gitPull はリモートリポジトリから取得します
func (s *Shell) gitPull(args []string) error {
	repo, err := s.getRepository()
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("ワークツリーの取得に失敗しました: %v", err)
	}

	err = worktree.Pull(&git.PullOptions{
		RemoteName: "origin",
	})

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			fmt.Println("既に最新の状態です")
			return nil
		}
		return fmt.Errorf("プルに失敗しました: %v", err)
	}

	fmt.Println("プルが完了しました")
	return nil
}

// gitLog はコミット履歴を表示します
func (s *Shell) gitLog(args []string) error {
	repo, err := s.getRepository()
	if err != nil {
		return err
	}

	ref, err := repo.Head()
	if err != nil {
		return fmt.Errorf("HEADの取得に失敗しました: %v", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return fmt.Errorf("ログの取得に失敗しました: %v", err)
	}

	count := 0
	maxCount := 10
	if len(args) > 0 && strings.Contains(args[0], "-") {
		// 簡単なオプション解析（-n 数値形式のみ対応）
		for i, arg := range args {
			if arg == "-n" && i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &maxCount)
				break
			}
		}
	}

	err = commitIter.ForEach(func(c *object.Commit) error {
		if count >= maxCount {
			return fmt.Errorf("達成")
		}
		fmt.Printf("%s %s\n", c.Hash.String()[:8], strings.Split(c.Message, "\n")[0])
		count++
		return nil
	})

	if err != nil && err.Error() != "達成" {
		return fmt.Errorf("ログの表示に失敗しました: %v", err)
	}
	return nil
}

// gitClone はリポジトリをクローンします
func (s *Shell) gitClone(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("エラー: クローンするリポジトリのURLを指定してください\n使用法: git clone https://github.com/user/repo.git")
	}

	url := args[0]
	var directory string

	if len(args) > 1 {
		directory = args[1]
	} else {
		// URLからリポジトリ名を抽出
		parts := strings.Split(url, "/")
		repoName := parts[len(parts)-1]
		if strings.HasSuffix(repoName, ".git") {
			repoName = repoName[:len(repoName)-4]
		}
		directory = repoName
	}

	// ディレクトリが既に存在するかチェック
	if _, err := os.Stat(directory); !os.IsNotExist(err) {
		return fmt.Errorf("エラー: ディレクトリ '%s' は既に存在します", directory)
	}

	fmt.Printf("リポジトリをクローンしています: %s -> %s\n", url, directory)

	_, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("クローンに失敗しました: %v", err)
	}

	fmt.Printf("クローンが完了しました: %s\n", directory)
	return nil
}

// gitHelp はGitコマンドのヘルプを表示します
func (s *Shell) gitHelp() {
	fmt.Println("利用可能なGitコマンド:")
	fmt.Println("  git status           - 作業ディレクトリの状態を表示")
	fmt.Println("  git add <ファイル>   - ファイルをステージングエリアに追加")
	fmt.Println("  git commit -m <メッセージ> - 変更をコミット")
	fmt.Println("  git push             - リモートリポジトリにプッシュ")
	fmt.Println("  git pull             - リモートリポジトリから取得")
	fmt.Println("  git log              - コミット履歴を表示")
	fmt.Println("  git clone <URL>      - リポジトリをクローン")
	fmt.Println("  git help             - このヘルプを表示")
}
