using System.Drawing;
using System.Windows.Forms;

namespace RfiPlugin
{
    /// <summary>最簡輸入視窗：標題＋說明＋回報人三欄。學生 0 經驗也看得懂。
    /// 右下角可拖拉放大；說明欄會跟著長高。</summary>
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
            Size = new Size(500, 500);
            MinimumSize = new Size(420, 400);
            StartPosition = FormStartPosition.CenterParent;
            // 可調大小＋右下角拖拉 grip（課上 demo 要大要小自己拉）
            FormBorderStyle = FormBorderStyle.Sizable;
            SizeGripStyle = SizeGripStyle.Show;
            MaximizeBox = false;
            MinimizeBox = false;
            Padding = new Padding(16);

            const int left = 20;
            const int width = 444;
            int y = 16;

            var lbl1 = new Label { Text = "標題（必填）", Top = y, Left = left, Width = width, Height = 28 };
            lbl1.Anchor = AnchorStyles.Top | AnchorStyles.Left | AnchorStyles.Right;
            y += 32;
            _title.Top = y; _title.Left = left; _title.Width = width; _title.Height = 34;
            _title.Anchor = AnchorStyles.Top | AnchorStyles.Left | AnchorStyles.Right;
            y += 48;

            var lbl2 = new Label { Text = "說明", Top = y, Left = left, Width = width, Height = 28 };
            lbl2.Anchor = AnchorStyles.Top | AnchorStyles.Left | AnchorStyles.Right;
            y += 32;
            int descTop = y;
            _desc.Top = descTop; _desc.Left = left; _desc.Width = width; _desc.Height = 150;
            _desc.Multiline = true; _desc.ScrollBars = ScrollBars.Vertical;
            _desc.Anchor = AnchorStyles.Top | AnchorStyles.Bottom | AnchorStyles.Left | AnchorStyles.Right;
            y += 166;

            var lbl3 = new Label { Text = "回報人", Top = y, Left = left, Width = width, Height = 28, Name = "lblReporter" };
            lbl3.Anchor = AnchorStyles.Bottom | AnchorStyles.Left | AnchorStyles.Right;
            y += 32;
            _reporter.Top = y; _reporter.Left = left; _reporter.Width = width; _reporter.Height = 34; _reporter.Name = "txtReporter";
            _reporter.Anchor = AnchorStyles.Bottom | AnchorStyles.Left | AnchorStyles.Right;
            y += 58;

            var ok = new Button { Text = "送出", DialogResult = DialogResult.OK, Top = y, Left = left + width - 200, Width = 96, Height = 36 };
            var cancel = new Button { Text = "取消", DialogResult = DialogResult.Cancel, Top = y, Left = left + width - 96, Width = 96, Height = 36 };
            ok.Anchor = AnchorStyles.Bottom | AnchorStyles.Right;
            cancel.Anchor = AnchorStyles.Bottom | AnchorStyles.Right;
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
