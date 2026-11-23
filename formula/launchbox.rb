class Launchbox < Formula
  # Homebrew formulas values are hardcoded and the reason
  # is due to how Homebrew’s DSL (domain-specific language in Ruby)
  # and its formula system are designed and parsed
  desc "Command-line interface for creating boilerplate Next.js application."
  homepage "https://github.com/fabiomrosa/homebrew-launchbox"
  license "MIT"
  version "0.7.2"

  on_macos do
    on_arm do
      url "https://github.com/fabiomrosa/homebrew-launchbox/releases/download/v#{version}/launchbox-darwin-arm64"
      sha256 "2a47b4f26e63b6d10928db5327280aa8bf8263dda52bbcb1dd194582a49b1805"
    end

    on_intel do
      url "https://github.com/fabiomrosa/homebrew-launchbox/releases/download/v#{version}/launchbox-darwin-amd64"
      sha256 "5dc7dd1822ca730a86956ef92f9d1805af5103d98dd61167b57355de66498f32"
    end
  end

  def install
    binary = if Hardware::CPU.arm?
      "launchbox-darwin-arm64"
    else
      "launchbox-darwin-amd64"
    end

    bin.install binary => "launchbox"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/launchbox --version")
  end
end
