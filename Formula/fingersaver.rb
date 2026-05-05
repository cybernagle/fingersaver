class Fingersaver < Formula
  desc "AI coding agent orchestrator with split-pane TUI"
  homepage "https://github.com/cybernagle/fingersaver"
  url "https://github.com/cybernagle/fingersaver/archive/refs/tags/v0.4.6.tar.gz"
  sha256 "96b5a308ac4e7da338568a7f885f0eb652211357823f0706d757076a0cf5b31f"
  version "0.4.6"

  depends_on "go" => :build
  depends_on "tmux"

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w")
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/fingersaver --version")
  end
end
