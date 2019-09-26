type Archive struct {
  jams []Gamejam

  m *model        // The model that holds the main archive data
  mPath []string  // The path in the (main) db to the archives
}

func NewArchive(m *model) *Archive {
  arc := new(Archive)
  arc.m = m
  arc.mPath = []string{"archive"}
  return arch
}

func (m *model) LoadArchive() (*Archive, error) {
  if err := m.openDB(); err != nil {
    return nil, err
  }
  defer m.closeDB()

  arc := NewArchive(m)
  vals, err := m.bolt.GetValueList(arc.mPath)
  if err != nil {
    // There apparently aren't any archived jams
    return arc, nil
  }
}

func (m *model) ArchiveCurrentJam() {

}
