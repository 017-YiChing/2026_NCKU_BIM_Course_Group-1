using System.Drawing;
using System.Windows.Forms;

namespace RfiPlugin
{
    /// <summary>最簡輸入視窗：標題＋說明＋回報人三欄。學生 0 經驗也看得懂。</summary>
    public class IssueForm : Form
    {
        private readonly TextBox _title = new TextBox();
        private readonly TextBox _desc = new TextBox();
        private readonly TextBox _reporter = new TextBox();

        public string IssueTitle => _title.Text.Trim();
        public string IssueDescription => _desc.Text.Trim();
        public string Reporter => _reporter.Text.Trim();

        public IssueForm()
        {
            Text = "建立 RFI 議題";
            Size = new Size(480, 460);
            StartPosition = FormStartPosition.CenterParent;
            FormBorderStyle = FormBorderStyle.FixedDialog;
            MaximizeBox = false;
            MinimizeBox = false;
            Padding = new Padding(16);

            const int left = 20;
            const int width = 424;
            int y = 16;

            var lbl1 = new Label { Text = "標題（必填）", Top = y, Left = left, Width = width, Height = 22 };
            y += 26;
            _title.Top = y; _title.Left = left; _title.Width = width; _title.Height = 28;
            y += 40;

            var lbl2 = new Label { Text = "說明", Top = y, Left = left, Width = width, Height = 22 };
            y += 26;
            _desc.Top = y; _desc.Left = left; _desc.Width = width; _desc.Height = 130;
            _desc.Multiline = true; _desc.ScrollBars = ScrollBars.Vertical;
            y += 146;

            var lbl3 = new Label { Text = "回報人", Top = y, Left = left, Width = width, Height = 22 };
            y += 26;
            _reporter.Top = y; _reporter.Left = left; _reporter.Width = width; _reporter.Height = 28;
            y += 52;

            var ok = new Button { Text = "送出", DialogResult = DialogResult.OK, Top = y, Left = left + width - 200, Width = 96, Height = 34 };
            var cancel = new Button { Text = "取消", DialogResult = DialogResult.Cancel, Top = y, Left = left + width - 96, Width = 96, Height = 34 };
            ok.Click += (s, e) =>
            {
                if (string.IsNullOrWhiteSpace(_title.Text))
                {
                    MessageBox.Show("標題必填");
                    DialogResult = DialogResult.None;
                }
            };

            Controls.AddRange(new Control[] { lbl1, _title, lbl2, _desc, lbl3, _reporter, ok, cancel });
            AcceptButton = ok;
            CancelButton = cancel;
        }
    }
}
