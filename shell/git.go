package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cherrysh/i18n"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
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

func (s *Shell) gitStatus(args []string) error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	status, err := worktree.Status()
	if err != nil {
		return err
	}

	if status.IsClean() {
		fmt.Println(i18n.T("git.clean_working_directory"))
		return nil
	}

	fmt.Println(i18n.T("git.changed_files"))

	// ステータスの種類を日本語で表示
	statusMap := map[git.StatusCode]string{
		git.Modified:   "変更",
		git.Added:      "追加",
		git.Deleted:    "削除",
		git.Renamed:    "名前変更",
		git.Copied:     "コピー",
		git.Untracked:  "未追跡",
		git.Unmodified: "未変更",
	}

	for file, fileStatus := range status {
		var statusTexts []string

		if fileStatus.Staging != git.Unmodified {
			if text, exists := statusMap[fileStatus.Staging]; exists {
				statusTexts = append(statusTexts, text)
			}
		}

		if fileStatus.Worktree != git.Unmodified {
			if text, exists := statusMap[fileStatus.Worktree]; exists {
				statusTexts = append(statusTexts, text)
			}
		}

		statusText := strings.Join(statusTexts, ", ")
		fmt.Printf("%s %s\n", statusText, file)
	}

	return nil
}

func (s *Shell) gitAdd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: git add <file>")
	}

	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	for _, file := range args {
		_, err := worktree.Add(file)
		if err != nil {
			fmt.Printf(i18n.T("git.add_error")+"\n", file, err)
			continue
		}
		fmt.Printf(i18n.T("git.add_success")+"\n", file)
	}

	return nil
}

func (s *Shell) gitCommit(args []string) error {
	if len(args) < 2 || args[0] != "-m" {
		return fmt.Errorf("usage: git commit -m \"message\"")
	}

	message := strings.Join(args[1:], " ")

	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	commit, err := worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Cherry Shell User",
			Email: "user@cherrysh.local",
		},
	})
	if err != nil {
		return err
	}

	fmt.Printf(i18n.T("git.commit_created")+"\n", commit.String()[:8])
	return nil
}

func (s *Shell) gitPush(args []string) error {
	// 外部のgitコマンドを使用
	cmd := exec.Command("git", append([]string{"push"}, args...)...)
	cmd.Dir = s.getCurrentDir()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println(i18n.T("git.already_up_to_date"))
	return nil
}

func (s *Shell) gitPull(args []string) error {
	// 外部のgitコマンドを使用
	cmd := exec.Command("git", append([]string{"pull"}, args...)...)
	cmd.Dir = s.getCurrentDir()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println(i18n.T("git.pull_completed"))
	return nil
}

func (s *Shell) gitLog(args []string) error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	ref, err := repo.Head()
	if err != nil {
		return err
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}

	// 最新の10件のコミットを表示
	count := 0
	maxCount := 10
	if len(args) > 0 && args[0] == "-n" && len(args) > 1 {
		// -n オプションで件数を指定
		fmt.Sscanf(args[1], "%d", &maxCount)
	}

	err = commitIter.ForEach(func(c *object.Commit) error {
		if count >= maxCount {
			return fmt.Errorf("stop iteration")
		}

		fmt.Printf("%s %s\n", c.Hash.String()[:8], strings.Split(c.Message, "\n")[0])
		count++
		return nil
	})

	if err != nil && err.Error() != "stop iteration" {
		return err
	}

	return nil
}

func (s *Shell) gitClone(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: git clone <url> [directory]")
	}

	url := args[0]
	directory := ""

	if len(args) > 1 {
		directory = args[1]
	} else {
		// URLからディレクトリ名を推測
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			directory = strings.TrimSuffix(parts[len(parts)-1], ".git")
		}
	}

	if directory == "" {
		return fmt.Errorf("could not determine directory name")
	}

	fmt.Printf(i18n.T("git.cloning_repository")+"\n", url, directory)

	// Clone options
	cloneOptions := &git.CloneOptions{
		URL: url,
	}

	// GitHub認証の設定
	if s.config != nil && s.config.GitHubToken != "" {
		// GitHubトークンを使用した認証
		cloneOptions.Auth = &http.BasicAuth{
			Username: s.config.GitHubUser, // GitHubの場合、ユーザー名は任意（トークンが重要）
			Password: s.config.GitHubToken,
		}
	}

	_, err := git.PlainClone(directory, false, cloneOptions)
	if err != nil {
		// 認証エラーの場合は適切なメッセージを表示
		if strings.Contains(err.Error(), "authentication required") {
			return fmt.Errorf("認証が必要です。.cherryshrcにGITHUB_TOKENを設定するか、SSH鍵を使用してください: %v", err)
		}
		return fmt.Errorf("クローンに失敗しました: %v", err)
	}

	fmt.Printf(i18n.T("git.clone_completed")+"\n", directory)
	return nil
}

func (s *Shell) gitHelp() {
	fmt.Println(i18n.T("git.help_title"))
	fmt.Println(i18n.T("git.help_status"))
	fmt.Println(i18n.T("git.help_add"))
	fmt.Println(i18n.T("git.help_commit"))
	fmt.Println(i18n.T("git.help_push"))
	fmt.Println(i18n.T("git.help_pull"))
	fmt.Println(i18n.T("git.help_log"))
	fmt.Println(i18n.T("git.help_clone"))
	fmt.Println(i18n.T("git.help_help"))
}
