export interface CsvColumn<T> {
  key: keyof T | string;
  header: string;
  format?: (val: any, row: T) => string;
}

export function exportToCsv<T extends Record<string, any>>(
  data: T[],
  filename: string,
  columns?: CsvColumn<T>[]
): void {
  if (!data || !data.length) {
    console.warn('exportToCsv: No data provided for export');
    return;
  }

  // Determine headers and accessors
  const cols: CsvColumn<T>[] = columns || Object.keys(data[0]).map(k => ({ key: k, header: k }));

  const csvRows: string[] = [];

  // Header row
  csvRows.push(cols.map(c => `"${escapeCsv(c.header)}"`).join(','));

  // Data rows
  for (const row of data) {
    const values = cols.map(c => {
      let val: any;
      if (typeof c.key === 'string' && c.key.includes('.')) {
        val = c.key.split('.').reduce((acc, part) => acc && acc[part], row);
      } else {
        val = row[c.key as keyof T];
      }

      if (c.format) {
        val = c.format(val, row);
      }

      if (val === null || val === undefined) {
        return '""';
      }

      return `"${escapeCsv(String(val))}"`;
    });
    csvRows.push(values.join(','));
  }

  const csvContent = 'data:text/csv;charset=utf-8,\uFEFF' + encodeURIComponent(csvRows.join('\r\n'));
  const link = document.createElement('a');
  link.setAttribute('href', csvContent);
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
  link.setAttribute('download', `${filename}_${timestamp}.csv`);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

function escapeCsv(str: string): string {
  return str.replace(/"/g, '""');
}
