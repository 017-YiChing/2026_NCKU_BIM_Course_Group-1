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
            Size = new Size(380, 300);
            StartPosition = FormStartPosition.CenterParent;
            FormBorderStyle = FormBorderStyle.FixedDialog;
            MaximizeBox = false;
            MinimizeBox = false;

            var lbl1 = new Label { Text = "標題（必填）", Top = 12, Left = 12, Width = 340 };
            _title.Top = 34; _title.Left = 12; _title.Width = 340;

            var lbl2 = new Label { Text = "說明", Top = 62, Left = 12, Width = 340 };
            _desc.Top = 84; _desc.Left = 12; _desc.Width = 340; _desc.Height = 80; _desc.Multiline = true;

            var lbl3 = new Label { Text = "回報人", Top = 172, Left = 12, Width = 340 };
            _reporter.Top = 194; _reporter.Left = 12; _reporter.Width = 340;

            var ok = new Button { Text = "送出", DialogResult = DialogResult.OK, Top = 226, Left = 196, Width = 75 };
            var cancel = new Button { Text = "取消", DialogResult = DialogResult.Cancel, Top = 226, Left = 279, Width = 75 };
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
