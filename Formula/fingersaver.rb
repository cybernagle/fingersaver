class Fingersaver < Formula
  desc "AI coding agent orchestrator with split-pane TUI"
  homepage "https://github.com/cybernagle/fingersaver"
  url "https://github.com/cybernagle/fingersaver/archive/refs/tags/v0.4.0.tar.gz"
  sha256 "480949bb5d0f0710e635a325d9759218b2ce69eb473c56659ae55460e0afb8ff"
  version "0.4.0"

  depends_on "go" => :build
  depends_on "tmux"

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w")
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/fingersaver --version")
  end
end
