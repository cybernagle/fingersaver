class Fingersaver < Formula
  desc "AI coding agent orchestrator with split-pane TUI"
  homepage "https://github.com/cybernagle/fingersaver"
  url "https://github.com/cybernagle/fingersaver/archive/refs/tags/v0.4.8.tar.gz"
  sha256 "e5cbd028738378a809d76c1e07f293712579883bd41125acea393ce5978695a5"
  version "0.4.8"

  depends_on "go" => :build
  depends_on "tmux"

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w")
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/fingersaver --version")
  end
end
