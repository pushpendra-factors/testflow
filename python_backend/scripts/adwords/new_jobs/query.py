import enum

class QueryBuilder():
    _select = None
    _from = None
    _where = None
    _orderby = None
    _limit = None

    def Select(self, fields):
        self._select = "SELECT "
        for i in range(len(fields)-1):
            self._select += fields[i] + ",  "
        self._select += fields[len(fields)-1]
        return self

    def From(self, _from):
        self._from = " FROM " + _from
        return self

    def Where(self, where):
        self._where = " WHERE " + where
        return self
    
    def During(self, during):
        if(self._where is None):
            self._where = " WHERE "
        else:
            self._where += " AND "
        self._where += "segments.date BETWEEN " + during
        return self

    def OrderBy(self, field):
        self._orderby = " ORDER BY " + field
        return self

    def Limit(self, limit):
        self._limit = " LIMIT " + str(limit)
        return self

    def Build(self):
        query = self._select
        query += self._from
        if(self._where is not None):
            query += self._where
        if(self._orderby is not None):
            query += self._orderby
        if(self._limit is not None):
            query += self._limit
        return query

    @staticmethod
    def getattribute(object, field):
        field_nested = field.split(".")
        for fl in field_nested:
            if hasattr(object, fl):
                object = getattr(object, fl)
            elif hasattr(object, '_'+fl):
                object = getattr(object, '_'+fl)
            elif hasattr(object, fl+'_'):
                object = getattr(object, fl+'_')
            else:
                return ""
        if isinstance(object, enum.Enum):
            object = str(int(object))
        return object