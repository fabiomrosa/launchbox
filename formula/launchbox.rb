class Launchbox < Formula
  # Homebrew formulas values are hardcoded and the reason
  # is due to how Homebrew’s DSL (domain-specific language in Ruby)
  # and its formula system are designed and parsed
  desc "Command-line interface for creating boilerplate Next.js application."
  homepage "https://github.com/fabiomrosa/homebrew-launchbox"
  license "MIT"
  version "0.7.1"

  on_macos do
    on_arm do
      url "https://github.com/fabiomrosa/homebrew-launchbox/releases/download/v#{version}/launchbox-darwin-arm64"
      sha256 "98ba209ed9cf9ac3b107a49eef945924e48c4d39df4824dccaff1f46d3dc0aea"
    end

    on_intel do
      url "https://github.com/fabiomrosa/homebrew-launchbox/releases/download/v#{version}/launchbox-darwin-amd64"
      sha256 "8ea57af1efe29df98f0f4ae14fe3ecadeafccd8931c7560dc1f682e1cbaa2899"
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
